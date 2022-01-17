package whatapi

import (
	"fmt"
	"html"
	"strconv"
)

type ArtistAlias struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	AliasID int    `json:"aliasid"`
}

type ExtendedArtistMap map[string][]ArtistAlias

type ArtistGroupStruct struct {
	GroupID               int                   `json:"groupId"`
	GroupYearF            int                   `json:"groupYear"`
	GroupRecordLabelF     string                `json:"groupRecordLabel"`
	GroupCatalogueNumberF string                `json:"groupCatalogueNumber"`
	GroupCategoryID       string                `json:"groupCategoryID"`
	TagsF                 []string              `json:"tags"`
	ReleaseTypeF          int                   `json:"releaseType"`
	GroupVanityHouse      bool                  `json:"groupVanityHouse"`
	HasBookmarked         bool                  `json:"hasBookmarked"`
	Torrent               []ArtistTorrentStruct `json:"torrent"`
	GroupNameF            string                `json:"groupName"`
	ArtistsF              []ArtistAlias         `json:"artists"`
	ExtendedArtists       ExtendedArtistMap     `json:"extendedArtists"`
	WikiImage             string                `json:"wikiImage"`
	artists               []string
	importance            []int
}

func (g ArtistGroupStruct) ID() int {
	return g.GroupID
}

func (g ArtistGroupStruct) Name() string {
	return html.UnescapeString(g.GroupNameF)
}

func (g ArtistGroupStruct) Artist() string {
	// this is embarrassing, we're embedded in an Artist object
	// we should be able to produce the Artist name

	// cut and pasted from r_torrent.go
	main := g.ExtendedArtists["1"]
	guest := g.ExtendedArtists["2"]
	remixer := g.ExtendedArtists["3"]
	composer := g.ExtendedArtists["4"]
	conductor := g.ExtendedArtists["5"]
	dj := g.ExtendedArtists["6"]
	producer := g.ExtendedArtists["7"]
	if ReleaseTypeString(g.ReleaseType()) == "Compilation" {
		if len(dj) == 1 {
			return html.UnescapeString(dj[0].Name)
		}
		if len(dj) == 2 {
			return fmt.Sprintf("%s & %s",
				html.UnescapeString(dj[0].Name),
				html.UnescapeString(dj[1].Name))
		}
	}
	if len(main) == 1 {
		return html.UnescapeString(main[0].Name)
	}
	if len(main) == 2 {
		return fmt.Sprintf("%s & %s",
			html.UnescapeString(main[0].Name),
			html.UnescapeString(main[1].Name))
	}
	if len(main) > 2 {
		return "VA"
	}
	// only if number of Artists == 0
	if len(composer) > 0 {
		return html.UnescapeString(composer[0].Name)
	}
	if len(guest) > 0 {
		return html.UnescapeString(guest[0].Name)
	}
	if len(remixer) > 0 {
		return html.UnescapeString(remixer[0].Name)
	}
	if len(conductor) > 0 {
		return html.UnescapeString(conductor[0].Name)
	}
	if len(producer) > 0 {
		return html.UnescapeString(producer[0].Name)
	}
	return "" // no name!
}

func (g ArtistGroupStruct) Year() int {
	return g.GroupYearF
}

func (g *ArtistGroupStruct) makeArtistsImportance() {
	g.artists = make([]string, 0, 7)
	g.importance = make([]int, 0, 7)
	for i := 1; i <= 7; i++ {
		for _, a := range g.ExtendedArtists[strconv.Itoa(i)] {
			g.artists = append(g.artists, a.Name)
			g.importance = append(g.importance, i)
		}
	}
}

func (g ArtistGroupStruct) Artists() []string {
	if g.artists == nil {
		g.makeArtistsImportance()
	}
	return g.artists
}

func (g ArtistGroupStruct) Importance() []int {
	if g.importance == nil {
		g.makeArtistsImportance()
	}
	return g.importance
}

func (g ArtistGroupStruct) RecordLabel() string {
	return html.UnescapeString(g.GroupRecordLabelF)
}

func (g ArtistGroupStruct) CatalogueNumber() string {
	return html.UnescapeString(g.GroupCatalogueNumberF)
}

func (g ArtistGroupStruct) ReleaseType() int {
	return g.ReleaseTypeF
}

func (g ArtistGroupStruct) Tags() []string {
	return g.TagsF
}

func (g ArtistGroupStruct) String() string {
	return GroupString(g)
}

type Artist struct {
	ID                   int    `json:"id"`
	NameF                string `json:"name"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
	HasBookmarked        bool   `json:"hasBookmarked"`
	Image                string `json:"image"`
	Body                 string `json:"body"`
	VanityHouse          bool   `json:"vanityHouse"`
	Tags                 []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
	SimilarArtists []struct {
		ArtistID  int    `json:"artistId"`
		Name      string `json:"name"`
		Score     int    `json:"score"`
		SimilarID int    `json:"similarId"`
	} `json:"similarArtists"`
	Statistics struct {
		NumGroups   int `json:"numGroups"`
		NumTorrents int `json:"numTorrents"`
		NumSeeders  int `json:"numSeeders"`
		NumLeechers int `json:"numLeechers"`
		NumSnatches int `json:"numSnatches"`
	} `json:"statistics"`
	TorrentGroup []ArtistGroupStruct `json:"torrentgroup"`
	Requests     []struct {
		RequestID  int    `json:"requestId"`
		CategoryID int    `json:"categoryId"`
		Title      string `json:"title"`
		Year       int    `json:"year"`
		TimeAdded  string `json:"timeAdded"`
		Votes      int    `json:"votes"`
		Bounty     int64  `json:"bounty"`
	} `json:"requests"`
}

func (a Artist) Name() string {
	return html.UnescapeString(a.NameF)
}
