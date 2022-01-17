package whatapi

import (
	"html"
)

type ArtistID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RequestsSearch struct {
	CurrentPage int `json:"currentPage"`
	Pages       int `json:"pages"`
	Results     []struct {
		RequestID       int          `json:"requestId"`
		RequestorID     int          `json:"requestorId"`
		ReqyestorName   string       `json:"requestorName"`
		TimeAdded       string       `json:"timeAdded"`
		LastVote        string       `json:"lastVote"`
		VoteCount       int          `json:"voteCount"`
		Bounty          int64        `json:"bounty"`
		CategoryID      int          `json:"categoryId"`
		CategoryName    string       `json:"categoryName"`
		Artists         [][]ArtistID `json:"artists"`
		Title           string       `json:"title"`
		Year            int          `json:"year"`
		Image           string       `json:"image"`
		Description     string       `json:"description"`
		CatalogueNumber string       `json:"catalogueNumber"`
		ReleaseType     string       `json:"releaseType"`
		BitrateList     string       `json:"bitrateList"`
		FormatList      string       `json:"formatList"`
		MediaList       string       `json:"mediaList"`
		LogCue          string       `json:"logCue"`
		IsFilled        bool         `json:"isFilled"`
		FillerID        int          `json:"fillerId"`
		FillerName      string       `json:"fillerName"`
		TorrentID       int          `json:"torrentId"`
		TimeFilled      string       `json:"timeFilled"`
	} `json:"results"`
}

type SearchTorrentStruct struct {
	TorrentID                int           `json:"torrentId"`
	EditionID                int           `json:"editionId"`
	Artists                  []ArtistAlias `json:"artists"`
	RemasteredF              bool          `json:"remastered"`
	RemasterYearF            int           `json:"remasterYear"`
	RemasterCatalogueNumberF string        `json:"remasterCatalogueNumber"`
	RemasterTitleF           string        `json:"remasterTitle"`
	MediaF                   string        `json:"media"`
	EncodingF                string        `json:"encoding"`
	FormatF                  string        `json:"format"`
	HasLogF                  bool          `json:"hasLog"`
	LogScore                 int           `json:"logScore"`
	HasCue                   bool          `json:"hasCue"`
	SceneF                   bool          `json:"scene"`
	VanityHouse              bool          `json:"vanityHouse"`
	FileCountF               int           `json:"fileCount"`
	Time                     string        `json:"time"`
	Size                     int64         `json:"size"`
	Snatches                 int           `json:"snatches"`
	Seeders                  int           `json:"seeders"`
	Leechers                 int           `json:"leechers"`
	IsFreeleech              bool          `json:"isFreeleech"`
	IsNeutralLeech           bool          `json:"isNeutralLeech"`
	IsPersonalFreeleech      bool          `json:"isPersonalFreeleech"`
	CanUseToken              bool          `json:"canUseToken"`
}

func (ts SearchTorrentStruct) ID() int {
	return ts.TorrentID
}

func (ts SearchTorrentStruct) Format() string {
	return ts.FormatF
}

func (ts SearchTorrentStruct) Encoding() string {
	return ts.EncodingF
}

func (ts SearchTorrentStruct) Media() string {
	return ts.MediaF
}

func (ts SearchTorrentStruct) Remastered() bool {
	return ts.RemasteredF
}

func (ts SearchTorrentStruct) RemasterCatalogueNumber() string {
	return ts.RemasterCatalogueNumberF
}

func (ts SearchTorrentStruct) RemasterTitle() string {
	return ts.RemasterTitleF
}

func (ts SearchTorrentStruct) RemasterYear() int {
	return ts.RemasterYearF
}

func (ts SearchTorrentStruct) Scene() bool {
	return ts.SceneF
}

func (ts SearchTorrentStruct) HasLog() bool {
	return ts.HasLogF
}

func (ts SearchTorrentStruct) String() string {
	return TorrentString(ts)
}

func (ts SearchTorrentStruct) FileCount() int {
	return ts.FileCountF
}

func (ts SearchTorrentStruct) FileSize() int64 {
	return ts.Size
}

type TorrentSearchResultStruct struct {
	GroupID       int                   `json:"groupId"`
	GroupName     string                `json:"groupName"`
	ArtistF       string                `json:"artist"`
	TagsF         []string              `json:"tags"`
	Bookmarked    bool                  `json:"bookmarked"`
	VanityHouse   bool                  `json:"vanityHouse"`
	GroupYear     int                   `json:"groupYear"`
	ReleaseTypeF  int                   `json:"releasetType,string"`
	GroupTime     string                `json:"groupTime"`
	TotalSnatched int                   `json:"totalSnatched"`
	TotalSeeders  int                   `json:"totalSeeders"`
	TotalLeechers int                   `json:"totalLeechers"`
	Torrents      []SearchTorrentStruct `json:"torrents"`
}

func (ts TorrentSearchResultStruct) ID() int {
	return ts.GroupID
}

func (ts TorrentSearchResultStruct) Name() string {
	return html.UnescapeString(ts.GroupName)
}

func (ts TorrentSearchResultStruct) Artist() string {
	return html.UnescapeString(ts.ArtistF)
}

func (ts TorrentSearchResultStruct) Year() int {
	return ts.GroupYear
}

func (ts TorrentSearchResultStruct) ReleaseType() int {
	return ts.ReleaseTypeF
}

func (ts TorrentSearchResultStruct) Tags() []string {
	return ts.TagsF
}

func (ts TorrentSearchResultStruct) String() string {
	return GroupString(ts)
}

type TorrentSearch struct {
	CurrentPage int                         `json:"currentPage"`
	Pages       int                         `json:"pages"`
	Results     []TorrentSearchResultStruct `json:"results"`
}

type UserSearch struct {
	CurrentPage int `json:"currentPage"`
	Pages       int `json:"pages"`
	Results     []struct {
		UserID   int    `json:"userId"`
		Username string `json:"username"`
		Donor    bool   `json:"donor"`
		Warned   bool   `json:"warned"`
		Enabled  bool   `json:"enabled"`
		Class    string `json:"class"`
	} `json:"results"`
}
