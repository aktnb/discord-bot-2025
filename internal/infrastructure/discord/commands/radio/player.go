package radio

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/bwmarrin/discordgo"
)

// playStream は FFmpeg を起動し、Radiko HLS ストリームを Discord ボイスに送信する。
// ctx がキャンセルされると FFmpeg を終了して返る。
func playStream(ctx context.Context, vc *discordgo.VoiceConnection, streamURL, authToken string) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-headers", fmt.Sprintf("X-Radiko-AuthToken: %s\r\n", authToken),
		"-i", streamURL,
		"-vn",
		"-c:a", "libopus",
		"-ar", "48000",
		"-ac", "2",
		"-b:a", "96k",
		"-application", "audio",
		"-f", "ogg",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("radio: stdout pipe: %v", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("radio: stderr pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("radio: ffmpeg start: %v", err)
		return
	}

	// stderr を破棄（ログノイズ防止）
	go io.Copy(io.Discard, stderr)

	defer func() {
		cmd.Wait()
		vc.Speaking(false)
	}()

	if err := vc.Speaking(true); err != nil {
		log.Printf("radio: speaking: %v", err)
		return
	}

	reader := &oggReader{r: stdout}
	packets := reader.readPackets(ctx)

	headerCount := 0
	for pkt := range packets {
		// オーディオ前の 2 つのヘッダーパケット（OpusHead, OpusTags）をスキップ
		if headerCount < 2 {
			headerCount++
			continue
		}
		select {
		case vc.OpusSend <- pkt:
		case <-ctx.Done():
			return
		}
	}
}

// oggReader は ogg コンテナから Opus パケットを読み出す
type oggReader struct {
	r io.Reader
}

// readPackets は ctx がキャンセルされるか読み込み終了まで Opus パケットを返すチャンネルを返す
func (or *oggReader) readPackets(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, 256)
	go func() {
		defer close(ch)
		var partial []byte
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			pkts, cont, err := or.readPage(partial)
			if err != nil {
				if err != io.EOF {
					log.Printf("radio: ogg read: %v", err)
				}
				return
			}
			partial = cont

			for _, p := range pkts {
				select {
				case ch <- p:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}

// readPage は ogg ページを 1 つ読み込み、完成したパケットのスライスと
// 次のページに持ち越す未完成パケットデータを返す
func (or *oggReader) readPage(partial []byte) (packets [][]byte, remaining []byte, err error) {
	// キャプチャパターン "OggS" を確認
	magic := make([]byte, 4)
	if _, err = io.ReadFull(or.r, magic); err != nil {
		return nil, nil, err
	}
	if string(magic) != "OggS" {
		return nil, nil, fmt.Errorf("invalid ogg magic: %x", magic)
	}

	// 固定ヘッダー: version(1) + header_type(1) + granule_pos(8) + serial(4) + seqno(4) + checksum(4) = 22 bytes
	header := make([]byte, 22)
	if _, err = io.ReadFull(or.r, header); err != nil {
		return nil, nil, err
	}

	// セグメント数
	nseg := make([]byte, 1)
	if _, err = io.ReadFull(or.r, nseg); err != nil {
		return nil, nil, err
	}

	// セグメントテーブル
	segTable := make([]byte, int(nseg[0]))
	if _, err = io.ReadFull(or.r, segTable); err != nil {
		return nil, nil, err
	}

	// ページデータを読み込む
	totalSize := 0
	for _, s := range segTable {
		totalSize += int(s)
	}
	data := make([]byte, totalSize)
	if _, err = io.ReadFull(or.r, data); err != nil {
		return nil, nil, err
	}

	// セグメントからパケットを組み立てる
	// セグメントサイズ < 255 でパケット終端
	current := append([]byte(nil), partial...)
	pos := 0
	for _, segSize := range segTable {
		current = append(current, data[pos:pos+int(segSize)]...)
		pos += int(segSize)
		if segSize < 255 {
			if len(current) > 0 {
				pkt := make([]byte, len(current))
				copy(pkt, current)
				packets = append(packets, pkt)
				current = nil
			}
		}
	}

	return packets, current, nil
}
