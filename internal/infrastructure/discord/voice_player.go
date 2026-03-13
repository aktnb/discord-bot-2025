package discord

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// oggPageHeader はOGGページのヘッダー情報
type oggPageHeader struct {
	headerType   byte
	nSegments    byte
	segmentTable []byte
}

// readOGGPagePackets はOGGページを1つ読み込み、含まれるパケット一覧を返す
// isHeader はこのページがヘッダーページかどうかを示す（BOS bit）
func readOGGPagePackets(r io.Reader) (packets [][]byte, isBOS bool, err error) {
	// OGGページの固定ヘッダー部分（27バイト）
	// 0-3:  capture_pattern "OggS"
	// 4:    stream_structure_version
	// 5:    header_type_flag
	// 6-13: granule_position (int64 LE)
	// 14-17: bitstream_serial_number
	// 18-21: page_sequence_no
	// 22-25: CRC_checksum
	// 26:   number_page_segments
	fixed := make([]byte, 27)
	if _, err = io.ReadFull(r, fixed); err != nil {
		return nil, false, err
	}
	if string(fixed[0:4]) != "OggS" {
		return nil, false, fmt.Errorf("OGGマジックが不正です")
	}

	headerType := fixed[5]
	isBOS = headerType&0x02 != 0
	nSegments := int(fixed[26])

	segTable := make([]byte, nSegments)
	if _, err = io.ReadFull(r, segTable); err != nil {
		return nil, false, err
	}

	// セグメントテーブルからパケットを組み立てる（Oggラッシング）
	var currentPacket []byte
	for _, segSize := range segTable {
		if segSize > 0 {
			buf := make([]byte, segSize)
			if _, err = io.ReadFull(r, buf); err != nil {
				return nil, false, err
			}
			currentPacket = append(currentPacket, buf...)
		}
		// セグメントサイズが255未満の場合はパケットの終端
		if segSize < 255 {
			if len(currentPacket) > 0 {
				packets = append(packets, currentPacket)
				currentPacket = nil
			}
		}
	}

	return packets, isBOS, nil
}

type playerSession struct {
	vc     *discordgo.VoiceConnection
	cancel context.CancelFunc
	done   chan struct{}
}

// VoicePlayer はボイスチャンネルでのラジオストリーミングを管理する
type VoicePlayer struct {
	session  *discordgo.Session
	mu       sync.Mutex
	sessions map[string]*playerSession // guildID -> session
}

// NewVoicePlayer はVoicePlayerを生成する
func NewVoicePlayer(s *discordgo.Session) *VoicePlayer {
	return &VoicePlayer{
		session:  s,
		sessions: make(map[string]*playerSession),
	}
}

// Play は指定したボイスチャンネルに参加してラジオのストリーミングを開始する
// 同じギルドで再生中の場合は停止してから新しい再生を開始する
func (vp *VoicePlayer) Play(guildID, channelID, streamURL, authToken string) error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// 既存セッションがあれば停止する
	if existing, ok := vp.sessions[guildID]; ok {
		existing.cancel()
		<-existing.done
		_ = existing.vc.Disconnect()
		delete(vp.sessions, guildID)
	}

	// ボイスチャンネルに参加する
	vc, err := vp.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("ボイスチャンネルへの参加に失敗しました: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	ps := &playerSession{vc: vc, cancel: cancel, done: done}
	vp.sessions[guildID] = ps

	go vp.streamAudio(ctx, ps, streamURL, authToken)

	return nil
}

// Stop は指定したギルドのラジオストリーミングを停止してボイスチャンネルから退出する
func (vp *VoicePlayer) Stop(guildID string) error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	ps, ok := vp.sessions[guildID]
	if !ok {
		return fmt.Errorf("再生中のラジオがありません")
	}

	ps.cancel()
	<-ps.done
	_ = ps.vc.Disconnect()
	delete(vp.sessions, guildID)

	return nil
}

// IsPlaying は指定したギルドでラジオが再生中かどうかを返す
func (vp *VoicePlayer) IsPlaying(guildID string) bool {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	_, ok := vp.sessions[guildID]
	return ok
}

// streamAudio はffmpegを使ってOGG/Opusストリームを取得し、DiscordへOpusパケットを送信する
func (vp *VoicePlayer) streamAudio(ctx context.Context, ps *playerSession, streamURL, authToken string) {
	defer close(ps.done)

	// ffmpegでHLSストリームをOGG/Opusに変換して標準出力に出力する
	args := []string{
		"-loglevel", "warning",
		"-headers", fmt.Sprintf("X-Radiko-AuthToken: %s\r\n", authToken),
		"-i", streamURL,
		"-c:a", "libopus",
		"-b:a", "96k",
		"-ar", "48000",
		"-ac", "2",
		"-f", "ogg",
		"pipe:1",
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[VoicePlayer] ffmpeg stdout pipe failed: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[VoicePlayer] ffmpeg start failed: %v", err)
		return
	}
	defer func() {
		_ = cmd.Wait()
	}()

	_ = ps.vc.Speaking(true)
	defer ps.vc.Speaking(false)

	// OGGストリームを読み込んでOpusパケットをDiscordに送信する
	// 最初の2ページ（OpusHead・OpusTags）はスキップする
	pageCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packets, _, err := readOGGPagePackets(stdout)
		if err != nil {
			if err != io.EOF && ctx.Err() == nil {
				log.Printf("[VoicePlayer] OGGページ読み込みエラー: %v", err)
			}
			return
		}

		pageCount++
		// 最初の2ページはOpusHead/OpusTagsヘッダーなのでスキップ
		if pageCount <= 2 {
			continue
		}

		for _, packet := range packets {
			if len(packet) == 0 {
				continue
			}
			select {
			case ps.vc.OpusSend <- packet:
			case <-ctx.Done():
				return
			}
		}
	}
}

