package whatapi

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

type GetTorrentStruct struct {
	Group   GroupStruct   `json:"group"`
	Torrent TorrentStruct `json:"torrent"`
}

type MusicInfoStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MusicInfo struct {
	Composers []MusicInfoStruct `json:"composers"`
	DJ        []MusicInfoStruct `json:"dj"`
	Artists   []MusicInfoStruct `json:"artists"`
	With      []MusicInfoStruct `json:"with"`
	Conductor []MusicInfoStruct `json:"conductor"`
	RemixedBy []MusicInfoStruct `json:"remixedBy"`
	Producer  []MusicInfoStruct `json:"producer"`
}

type GroupStruct struct {
	WikiBodyF        string    `json:"wikiBody"`
	WikiImageF       string    `json:"wikiImage"`
	IDF              int       `json:"id"`
	NameF            string    `json:"name"`
	YearF            int       `json:"year"`
	RecordLabelF     string    `json:"recordLabel"`
	CatalogueNumberF string    `json:"catalogueNumber"`
	ReleaseTypeF     int       `json:"releaseType"`
	CategoryID       int       `json:"categoryId"`
	CategoryName     string    `json:"categoryName"`
	Time             string    `json:"time"`
	VanityHouse      bool      `json:"vanityHouse"`
	IsBookmarked     bool      `json:"isBookmarked"`
	MusicInfo        MusicInfo `json:"musicInfo"`
	TagsF            []string  `json:"tags"`
	artists          []string
	importance       []int
}

type GroupStructTagsID struct {
	WikiBodyF        string         `json:"wikiBody"`
	WikiImageF       string         `json:"wikiImage"`
	IDF              int            `json:"id"`
	NameF            string         `json:"name"`
	YearF            int            `json:"year"`
	RecordLabelF     string         `json:"recordLabel"`
	CatalogueNumberF string         `json:"catalogueNumber"`
	ReleaseTypeF     int            `json:"releaseType"`
	CategoryID       int            `json:"categoryId"`
	CategoryName     string         `json:"categoryName"`
	Time             string         `json:"time"`
	VanityHouse      bool           `json:"vanityHouse"`
	IsBookmarked     bool           `json:"isBookmarked"`
	MusicInfo        MusicInfo      `json:"musicInfo"`
	TagsF            map[int]string `json:"tags"`
	artists          []string
	importance       []int
}

func (g GroupStruct) ID() int {
	return g.IDF
}

func (g GroupStruct) Name() string {
	return html.UnescapeString(g.NameF)
}

func (g GroupStruct) Artist() string {
	if ReleaseTypeString(g.ReleaseType()) == "Compilation" {
		if len(g.MusicInfo.DJ) == 1 {
			return html.UnescapeString(g.MusicInfo.DJ[0].Name)
		}
		if len(g.MusicInfo.DJ) == 2 {
			return fmt.Sprintf("%s & %s",
				html.UnescapeString(g.MusicInfo.DJ[0].Name),
				html.UnescapeString(g.MusicInfo.DJ[1].Name))
		}
	}
	if len(g.MusicInfo.Artists) == 1 {
		return html.UnescapeString(g.MusicInfo.Artists[0].Name)
	}
	if len(g.MusicInfo.Artists) == 2 {
		return fmt.Sprintf("%s & %s",
			html.UnescapeString(g.MusicInfo.Artists[0].Name),
			html.UnescapeString(g.MusicInfo.Artists[1].Name))
	}
	if len(g.MusicInfo.Artists) > 2 {
		return "VA"
	}
	// only if number of Artists == 0
	if len(g.MusicInfo.Composers) > 0 {
		return html.UnescapeString(g.MusicInfo.Composers[0].Name)
	}
	if len(g.MusicInfo.With) > 0 {
		return html.UnescapeString(g.MusicInfo.With[0].Name)
	}
	if len(g.MusicInfo.RemixedBy) > 0 {
		return html.UnescapeString(g.MusicInfo.RemixedBy[0].Name)
	}
	if len(g.MusicInfo.Conductor) > 0 {
		return html.UnescapeString(g.MusicInfo.Conductor[0].Name)
	}
	if len(g.MusicInfo.Producer) > 0 {
		return html.UnescapeString(g.MusicInfo.Producer[0].Name)
	}
	return "" // no name!
}

func (g GroupStruct) Year() int {
	return g.YearF
}

func (g GroupStruct) WikiImage() string {
	return g.WikiImageF
}

func (g *GroupStruct) makeArtistsImportance() {
	g.artists = make([]string, 0, len(g.MusicInfo.Composers)+
		len(g.MusicInfo.DJ)+
		len(g.MusicInfo.Artists)+
		len(g.MusicInfo.With)+
		len(g.MusicInfo.Conductor)+
		len(g.MusicInfo.RemixedBy)+
		len(g.MusicInfo.Producer))
	g.importance = make([]int, 0, len(g.artists))
	for _, n := range g.MusicInfo.Composers {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 4)
	}
	for _, n := range g.MusicInfo.DJ {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 6)
	}
	for _, n := range g.MusicInfo.Artists {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 1)
	}
	for _, n := range g.MusicInfo.With {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 2)
	}
	for _, n := range g.MusicInfo.Conductor {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 5)
	}
	for _, n := range g.MusicInfo.RemixedBy {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 3)
	}
	for _, n := range g.MusicInfo.Producer {
		g.artists = append(g.artists, n.Name)
		g.importance = append(g.importance, 7)
	}
}

func (g GroupStruct) Artists() []string {
	if g.artists == nil {
		g.makeArtistsImportance()
	}
	return g.artists
}

func (g GroupStruct) Importance() []int {
	if g.importance == nil {
		g.makeArtistsImportance()
	}
	return g.importance
}

func (g GroupStruct) RecordLabel() string {
	return g.RecordLabelF
}

func (g GroupStruct) CatalogueNumber() string {
	return html.UnescapeString(g.CatalogueNumberF)
}

func (g GroupStruct) ReleaseType() int {
	return g.ReleaseTypeF
}

func (g GroupStruct) Tags() []string {
	return g.TagsF
}

func (g GroupStruct) WikiBody() string {
	return g.WikiBodyF
}

func (g GroupStruct) String() string {
	return GroupString(g)
}

/* torrent from top10
{
    "artist": "Ty Segall",
    "data": 4562327934,
    "encoding": "Lossless",
    "format": "FLAC",
    "groupCategory": 1,
    "groupId": 498588,
    "groupName": "Deforming Lobes",
    "groupYear": 2019,
    "hasCue": true,
    "hasLog": true,
    "hasLogDB": true,
    "leechers": 0,
    "logChecksum": "1",
    "logScore": "100",
    "media": "CD",
    "releaseType": "11"
    "remasterTitle": "",
    "scene": false,
    "seeders": 15,
    "size": 253462663,
    "snatched": 18,
    "tags": ["rock","alternative.rock"],
    "torrentId": 1110926,
    "wikiImage": "https://ptpimg.me/3952mr.jpg",
    "year": 2019,
}
*/
/* torrent from search/browse
{
    "artists": [{"id": 17521,"name": "Budapest Festival Orchestra","aliasid": 17521}],
    "canUseToken": true,
    "editionId": 1,
    "encoding": "Lossless",
    "fileCount": 10,
    "format": "FLAC",
    "hasCue": true,
    "hasLog": true,
    "hasSnatched": false
    "isFreeleech": false,
    "isNeutralLeech": false,
    "isPersonalFreeleech": false,
    "leechers": 0,
    "logScore": 90,
    "media": "CD",
    "remasterCatalogueNumber": "",
    "remasterTitle": "",
    "remasterYear": 0,
    "remastered": false,
    "scene": false,
    "seeders": 10,
    "size": 193280644,
    "snatches": 0,
    "time": "2016-12-01 13:40:41",
    "torrentId": 109795,
    "vanityHouse": false,
}
*/
/* torrent from artist
{
    "encoding": "24bit Lossless",
    "fileCount": 8,
    "format": "FLAC",
    "freeTorrent": false,
    "groupId": 492847,
    "hasCue": false,
    "hasFile": 1096610
    "hasLog": false,
    "id": 1096610,
    "leechers": 0,
    "logScore": 100,
    "media": "WEB",
    "remasterRecordLabel": "Channel Classics",
    "remasterTitle": "",
    "remasterYear": 2019,
    "remastered": true,
    "scene": false,
    "seeders": 6,
    "size": 3867570202,
    "snatched": 5,
    "time": "2019-03-09 06:26:21",
}
*/
/* torrent from torrent
{
    "id": 2281083,
    "infoHash": "0108C105006D386A44D8C0288603C52873F1E40F",
    "media": "CD",
    "format": "FLAC",
    "encoding": "Lossless",
    "remastered": true,
    "remasterYear": 2019,
    "remasterTitle": "",
    "remasterRecordLabel": "Sub Pop Records",
    "remasterCatalogueNumber": "SP1232",
    "scene": false,
    "hasLog": true,
    "hasCue": true,
    "logScore": 100,
    "fileCount": 21,
    "size": 253867654,
    "seeders": 90,
    "leechers": 0,
    "snatched": 82,
    "freeTorrent": false,
    "reported": false,
    "time": "2019-04-05 03:11:57",
    "description": "From retail CD. 200 dpi scans included.",
    "fileList": "01 - A Lot&#39;s Gonna Change.flac{{{25192410}}}|||02 - Andromeda.flac{{{29215614}}}|||03 - Everyday.flac{{{31396102}}}|||04 - Something to Believe.flac{{{29275146}}}|||05 - Titanic Rising.flac{{{7218613}}}|||06 - Movies.flac{{{34746766}}}|||07 - Mirror Forever.flac{{{30477537}}}|||08 - Wild Time.flac{{{38157285}}}|||09 - Picture Me Better.flac{{{20285291}}}|||10 - Nearer to Thee.flac{{{5569561}}}|||Artwork/back.jpg{{{388980}}}|||Artwork/disc.jpg{{{304088}}}|||Artwork/front.jpg{{{377827}}}|||Artwork/inside-l.jpg{{{373471}}}|||Artwork/inside-r.jpg{{{420274}}}|||Artwork/spine.jpg{{{22753}}}|||Artwork/sticker.jpg{{{76359}}}|||Titanic Rising.cue{{{1926}}}|||Weyes Blood - Titanic Rising.jpg{{{354900}}}|||Weyes Blood - Titanic Rising.log{{{12082}}}|||Weyes Blood - Titanic Rising.m3u{{{669}}}",
    "filePath": "Weyes Blood - Titanic Rising (2019) [FLAC] {SP 1232}",
    "userId": 2661,
    "username": "rogueofmv"
}
*/
/* torrent from torrentgroup
{
    "id": 2281083,
    "media": "CD",
    "format": "FLAC",
    "encoding": "Lossless",
    "remastered": true,
    "remasterYear": 2019,
    "remasterTitle": "",
    "remasterRecordLabel": "Sub Pop Records",
    "remasterCatalogueNumber": "SP1232",
    "scene": false,
    "hasLog": true,
    "hasCue": true,
    "logScore": 100,
    "fileCount": 21,
    "size": 253867654,
    "seeders": 90,
    "leechers": 0,
    "snatched": 82,
    "freeTorrent": false,
    "reported": false,
    "time": "2019-04-05 03:11:57",
    "description": "From retail CD. 200 dpi scans included.",
    "fileList": "01 - A Lot&#39;s Gonna Change.flac{{{25192410}}}|||02 - Andromeda.flac{{{29215614}}}|||03 - Everyday.flac{{{31396102}}}|||04 - Something to Believe.flac{{{29275146}}}|||05 - Titanic Rising.flac{{{7218613}}}|||06 - Movies.flac{{{34746766}}}|||07 - Mirror Forever.flac{{{30477537}}}|||08 - Wild Time.flac{{{38157285}}}|||09 - Picture Me Better.flac{{{20285291}}}|||10 - Nearer to Thee.flac{{{5569561}}}|||Artwork/back.jpg{{{388980}}}|||Artwork/disc.jpg{{{304088}}}|||Artwork/front.jpg{{{377827}}}|||Artwork/inside-l.jpg{{{373471}}}|||Artwork/inside-r.jpg{{{420274}}}|||Artwork/spine.jpg{{{22753}}}|||Artwork/sticker.jpg{{{76359}}}|||Titanic Rising.cue{{{1926}}}|||Weyes Blood - Titanic Rising.jpg{{{354900}}}|||Weyes Blood - Titanic Rising.log{{{12082}}}|||Weyes Blood - Titanic Rising.m3u{{{669}}}",
    "filePath": "Weyes Blood - Titanic Rising (2019) [FLAC] {SP 1232}",
    "userId": 2661,
    "username": "rogueofmv"
}
*/

type ArtistTorrentStruct struct {
	IDF                  int    `json:"id"`
	GroupIDF             int    `json:"groupId"`
	MediaF               string `json:"media"`
	FormatF              string `json:"format"`
	EncodingF            string `json:"encoding"`
	RemasterYearF        int    `json:"remasterYear"`
	RemasteredF          bool   `json:"remastered"`
	RemasterTitleF       string `json:"remasterTitle"`
	RemasterRecordLabelF string `json:"remasterRecordLabel"`
	SceneF               bool   `json:"scene"`
	HasLogF              bool   `json:"hasLog"`
	HasCue               bool   `json:"hasCue"`
	LogScore             int    `json:"logScore"`
	FileCountF           int    `json:"fileCount"`
	FreeTorrent          string `json:"freeTorrent"` // actually bool
	Size                 int64  `json:"size"`
	Leechers             int    `json:"leechers"`
	Seeders              int    `json:"seeders"`
	Snatched             int    `json:"snatched"`
	Time                 string `json:"time"`
	HasFile              int    `json:"hasFile"`
}

func (t ArtistTorrentStruct) ID() int {
	return t.IDF
}

func (t ArtistTorrentStruct) GroupID() int {
	return t.GroupIDF
}

func (t ArtistTorrentStruct) Format() string {
	return t.FormatF
}

func (t ArtistTorrentStruct) Encoding() string {
	return t.EncodingF
}

func (t ArtistTorrentStruct) Media() string {
	return t.MediaF
}

func (t ArtistTorrentStruct) Remastered() bool {
	return t.RemasteredF
}

func (t ArtistTorrentStruct) RemasterRecordLabel() string {
	return html.UnescapeString(t.RemasterRecordLabelF)
}

func (t ArtistTorrentStruct) RemasterTitle() string {
	return html.UnescapeString(t.RemasterTitleF)
}

func (t ArtistTorrentStruct) RemasterYear() int {
	return t.RemasterYearF
}
func (t ArtistTorrentStruct) Scene() bool {
	return t.SceneF
}
func (t ArtistTorrentStruct) HasLog() bool {
	return t.HasLogF
}
func (t ArtistTorrentStruct) String() string {
	return TorrentString(t)
}
func (t ArtistTorrentStruct) FileCount() int {
	return t.FileCountF
}
func (t ArtistTorrentStruct) FileSize() int64 {
	return t.Size
}

type TorrentStruct struct {
	IDF                      int    `json:"id"`
	InfoHash                 string `json:"infoHash"`
	MediaF                   string `json:"media"`
	FormatF                  string `json:"format"`
	EncodingF                string `json:"encoding"`
	RemasteredF              bool   `json:"remastered"`
	RemasterYearF            int    `json:"remasterYear"`
	RemasterTitleF           string `json:"remasterTitle"`
	RemasterRecordLabelF     string `json:"remasterRecordLabel"`
	RemasterCatalogueNumberF string `json:"remasterCatalogueNumber"`
	SceneF                   bool   `json:"scene"`
	HasLogF                  bool   `json:"hasLog"`
	HasCue                   bool   `json:"hasCue"`
	LogScore                 int    `json:"logScore"`
	FileCountF               int    `json:"fileCount"`
	Size                     int64  `json:"size"`
	Seeders                  int    `json:"seeders"`
	Leechers                 int    `json:"leechers"`
	Snatched                 int    `json:"snatched"`
	FreeTorrent              string `json:"freeTorrent"` // actually bool
	Reported                 bool   `json:"reported"`
	Time                     string `json:"time"`
	DescriptionF             string `json:"description"`
	FileList                 string `json:"fileList"`
	FilePathF                string `json:"filePath"`
	UserID                   int    `json:"userID"`
	Username                 string `json:"username"`
	files                    []FileStruct
}

func (t TorrentStruct) ID() int {
	return t.IDF
}

func (t TorrentStruct) Format() string {
	return t.FormatF
}

func (t TorrentStruct) Encoding() string {
	return t.EncodingF
}

func (t TorrentStruct) Media() string {
	return t.MediaF
}

func (t TorrentStruct) Remastered() bool {
	return t.RemasteredF
}

func (t TorrentStruct) RemasterRecordLabel() string {
	return html.UnescapeString(t.RemasterRecordLabelF)
}

func (t TorrentStruct) RemasterCatalogueNumber() string {
	return html.UnescapeString(t.RemasterCatalogueNumberF)
}

func (t TorrentStruct) RemasterTitle() string {
	return html.UnescapeString(t.RemasterTitleF)
}

func (t TorrentStruct) RemasterYear() int {
	return t.RemasterYearF
}
func (t TorrentStruct) Description() string {
	return t.DescriptionF
}
func (t TorrentStruct) Scene() bool {
	return t.SceneF
}
func (t TorrentStruct) HasLog() bool {
	return t.HasLogF
}
func (t TorrentStruct) String() string {
	return TorrentString(t)
}
func (t TorrentStruct) FilePath() string {
	return html.UnescapeString(t.FilePathF)
}
func (t TorrentStruct) FileCount() int {
	return t.FileCountF
}
func (t TorrentStruct) FileSize() int64 {
	return t.Size
}
func (t *TorrentStruct) Files() (files []FileStruct, err error) {
	files, parse_err := t.parseFileList()
	if parse_err != nil {
		return files, parse_err
	}
	t.files = files
	return t.files, nil
}

// FileStruct represents what we know about the files in a torrent
type FileStruct struct {
	NameF string
	Size  int64
}

func (fs FileStruct) Name() string {
	return html.UnescapeString(fs.NameF)
}

// ParseFileList returns a slice of FileStruts for a torrent
func (t TorrentStruct) parseFileList() ([]FileStruct, error) {
	if t.FileList == "" {
		return []FileStruct{}, nil
	}
	f := []FileStruct{}
	re := regexp.MustCompile("(.*){{{(.*)}}}")
	for _, s := range strings.Split(t.FileList, "|||") {
		m := re.FindStringSubmatch(s)
		if len(m) != 3 {
			return nil, fmt.Errorf("could not parse %s", s)
		}
		i, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse %s: %s", s, err)
		}
		f = append(f, FileStruct{m[1], i})
	}
	return f, nil
}
