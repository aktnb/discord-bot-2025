package radiko

// Station はラジオ局を表す
type Station struct {
	ID   string
	Name string
}

// PredefinedStations はラジオ局の一覧
var PredefinedStations = []Station{
	{ID: "TBS", Name: "TBSラジオ"},
	{ID: "QRR", Name: "文化放送"},
	{ID: "LFR", Name: "ニッポン放送"},
	{ID: "JORF", Name: "ラジオ日本"},
	{ID: "JOAY", Name: "NHKラジオ第1(東京)"},
	{ID: "JOBY", Name: "NHKラジオ第2"},
	{ID: "JOFR", Name: "NHK FM(東京)"},
	{ID: "FMT", Name: "TOKYO FM"},
	{ID: "FMJ", Name: "J-WAVE"},
	{ID: "NACK5", Name: "NACK5"},
	{ID: "YFM", Name: "横浜FM"},
	{ID: "INT", Name: "InterFM897"},
	{ID: "MBS", Name: "MBSラジオ"},
	{ID: "ABC", Name: "ABCラジオ"},
	{ID: "OBC", Name: "ラジオ大阪"},
}

// FindStation はIDでラジオ局を検索する
func FindStation(id string) (Station, bool) {
	for _, s := range PredefinedStations {
		if s.ID == id {
			return s, true
		}
	}
	return Station{}, false
}
