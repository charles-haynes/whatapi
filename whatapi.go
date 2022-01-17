//Package whatapi is a wrapper for the What.CD JSON API (https://github.com/WhatCD/Gazelle/wiki/JSON-API-Documentation).
package whatapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
	retrywait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

type PSList struct{}

func (p PSList) PublicSuffix(d string) string {
	s, err := publicsuffix.EffectiveTLDPlusOne(d)
	if err != nil {
		return d
	}
	return s
}

func (p PSList) String() string {
	return "whatapi PSList v1.0"
}

//NewClient creates a new client for the What.CD API using the provided URL.
func NewClient(ur, agent string) (Client, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(ur)
	if err != nil {
		return nil, err
	}
	return &ClientStruct{
		baseURL:   *u,
		userAgent: agent,
		client:    &http.Client{Jar: cookieJar},
		db:        nil,
		cacheFor:  0,
	}, nil
}

// Cache caches requests and responses from a What.CD API client using
// the provided sql db as a cache. It returns cached responses newer
// than the cacheFor duration. It initialises the cache if needed.
func Cache(whatAPI Client, db *sql.DB, cacheFor time.Duration) (Client, error) {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS urlcache (
    requesturl TEXT PRIMARY KEY NOT NULL,
    body       TEXT NOT NULL,
    timestamp  DATETIME NOT NULL
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS cookies (
    url    TEXT PRIMARY KEY NOT NULL,
    cookie TEXT NOT NULL
) WITHOUT ROWID;
`)
	if err != nil {
		return nil, err
	}
	w, ok := whatAPI.(*ClientStruct)
	if !ok {
		return nil,
			fmt.Errorf("can only wrap ClientStruct at this time")
	}
	wCopy := *w
	wCopy.db = db
	wCopy.cacheFor = cacheFor
	return &wCopy, nil
}

type Group interface {
	ID() int
	Name() string
	Artist() string
	Year() int
	ReleaseType() int
	Tags() []string
	String() string
}

func ReleaseTypeString(r int) string {
	s := map[int]string{
		1:  "Album",
		3:  "Soundtrack",
		5:  "EP",
		6:  "Anthology",
		7:  "Compilation",
		9:  "Single",
		11: "Live",
		13: "Remix",
		14: "Bootleg",
		15: "Interview",
		16: "Mixtape",
		17: "Demo",
		18: "Concert",
		19: "DJ",
		21: "Unknown",
	}
	if v, ok := s[r]; ok {
		return v
	}
	return "Invalid Release Type"
}

type GroupRelease interface {
	Group
	RecordLabel() string
	CatalogueNumber() string
}

type GroupExt interface {
	GroupRelease
	WikiImage() string
	Artists() []string
	Importance() []int
	WikiBody() string
}

func oneOrTwoMusicInfos(mi []MusicInfoStruct) string {
	switch len(mi) {
	case 1:
		return mi[0].Name
	case 2:
		return fmt.Sprintf("%s & %s", mi[0].Name, mi[1].Name)
	default:
		return ""
	}
}

func GroupString(g Group) string {
	for _, t := range g.Tags() {
		if t != "classical" {
			continue
		}
		gs, ok := g.(GroupStruct)
		if !ok {
			break
		}
		mi := gs.MusicInfo
		s := []string{}
		if i := oneOrTwoMusicInfos(mi.Composers); i != "" {
			s = append(s, i, "-")
		}
		s = append(s, gs.Name(), "-")
		if i := oneOrTwoMusicInfos(mi.Artists); i != "" {
			s = append(s, i)
		}
		if i := oneOrTwoMusicInfos(mi.Conductor); i != "" {
			s = append(s, i)
		}
		s = append(s, fmt.Sprintf("(%4d)", gs.Year()))
		return strings.Join(s, " ")
	}
	return fmt.Sprintf("%s - %s (%4d)", g.Artist(), g.Name(), g.Year())
}

type Torrent interface {
	ID() int
	Format() string
	Encoding() string
	Media() string
	Remastered() bool
	RemasterTitle() string
	RemasterYear() int
	Scene() bool
	HasLog() bool
	String() string
	FileCount() int
	FileSize() int64
}

type TorrentCatalogueNumber interface {
	RemasterCatalogueNumber() string
}

type TorrentRecordLabel interface {
	RemasterRecordLabel() string
}

type TorrentFiles interface {
	Torrent
	TorrentRecordLabel
	TorrentCatalogueNumber
	FilePath() string
	Files() ([]FileStruct, error)
}

type TorrentExt interface {
	TorrentFiles
	Description() string
}

func TorrentString(t Torrent) string {
	s := fmt.Sprintf("[%s %s %s]", t.Media(), t.Format(), t.Encoding())
	if !t.Remastered() {
		return s
	}
	label := ""
	if r, ok := t.(TorrentRecordLabel); ok {
		label = r.RemasterRecordLabel()
	}
	number := ""
	if r, ok := t.(TorrentCatalogueNumber); ok {
		number = r.RemasterCatalogueNumber()
	}
	s = fmt.Sprintf("%s{(%4d) %s/%s/%s}",
		s, t.RemasterYear(), label,
		number, t.RemasterTitle())
	return s
}

//Client represents a client for the What.CD API.
type Client interface {
	GetJSON(requestURL string, responseObj interface{}) error
	Do(action string, params url.Values, result interface{}) error
	CreateDownloadURL(id int) (string, error)
	CreateUploadURL() (url.URL, string, error)
	Login(username, password string) error
	Logout() error
	GetAccount() error
	GetMailbox(params url.Values) (Mailbox, error)
	GetConversation(id int) (Conversation, error)
	GetNotifications(params url.Values) (Notifications, error)
	GetAnnouncements() (Announcements, error)
	GetSubscriptions(params url.Values) (Subscriptions, error)
	GetCategories() (Categories, error)
	GetForum(id int, params url.Values) (Forum, error)
	GetThread(id int, params url.Values) (Thread, error)
	GetArtistBookmarks() (ArtistBookmarks, error)
	GetTorrentBookmarks() (TorrentBookmarks, error)
	GetArtist(id int, params url.Values) (Artist, error)
	GetRequest(id int, params url.Values) (Request, error)
	GetTorrent(id int, params url.Values) (GetTorrentStruct, error)
	GetTorrentGroup(id int, params url.Values) (TorrentGroup, error)
	SearchTorrents(searchStr string, params url.Values) (TorrentSearch, error)
	SearchRequests(searchStr string, params url.Values) (RequestsSearch, error)
	SearchUsers(searchStr string, params url.Values) (UserSearch, error)
	GetTopTenTorrents(params url.Values) (TopTenTorrents, error)
	GetTopTenTags(params url.Values) (TopTenTags, error)
	GetTopTenUsers(params url.Values) (TopTenUsers, error)
	GetSimilarArtists(id, limit int) (SimilarArtists, error)
}

//ClientStruct represents a client for the What.CD API.
type ClientStruct struct {
	baseURL   url.URL
	userAgent string
	client    *http.Client
	authkey   string
	passkey   string
	loggedIn  bool
	db        *sql.DB
	cacheFor  time.Duration
}

// Client gets the http client for low level requests
func (w ClientStruct) Client() *http.Client {
	return w.client
}

// doRequest exectutes an http.Request on this server and returns the results
// or an error if the response was anything except 200
func (w *ClientStruct) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", w.userAgent)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errRequestFailedReason("Status Code " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (w *ClientStruct) updateCache(requestURL string, body []byte) error {
	if w.db == nil {
		return nil
	}
	res, err := w.db.Exec(
		"REPLACE INTO urlcache (requesturl, body, timestamp) "+
			"VALUES(?,?, datetime('now'))",
		requestURL, body)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf(
			"INSERT affected %d rows, expected 1", rows)
	}
	return nil
}

func (w *ClientStruct) cachedResponse(requestURL string) (body []byte, err error) {
	if w.db == nil {
		return nil, nil
	}

	var timestamp time.Time
	err = w.db.QueryRow(
		"SELECT body, timestamp FROM urlcache WHERE requesturl = ?", requestURL).
		Scan(&body, &timestamp)
	if err != nil {
		return nil, err
	}
	if body == nil || len(body) == 0 || time.Since(timestamp) > w.cacheFor {
		return nil, sql.ErrNoRows
	}
	return body, err
}

//GetJSON sends a HTTP GET request to the API and decodes the JSON response into responseObj.
func (w *ClientStruct) GetJSON(requestURL string, responseObj interface{}) (err error) {
	if !w.loggedIn {
		return errRequestFailedLogin
	}

	body, err := w.cachedResponse(requestURL)
	switch {
	case w.db == nil || err == sql.ErrNoRows:
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return err
		}

		err = retry.OnError(retrywait.Backoff{
			Duration: 15 * time.Second,
			Steps:    4,
			Factor:   3,
			Jitter:   0.1,
		}, func(err error) bool {
			es := err.Error()
			return es == "Request failed: Status Code 503 Service Unavailable" || // err from html
				es == "Request failed: Rate limit exceeded" // err from json
		}, func() (err error) {
			if body, err = w.doRequest(req); err != nil {
				return err // err from http
			}

			var st GenericResponse
			if err := json.Unmarshal(body, &st); err != nil {
				return err
			}
			if err := checkResponseStatus(st.Status, st.Error); err != nil {
				return err // err from json
			}

			return err
		})
		if err != nil {
			return err
		}

		if err = w.updateCache(requestURL, body); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		break
	}

	switch ro := responseObj.(type) {
	case *ArtistResponse: // hack around orpheus bug in get artist
		err := json.Unmarshal(body, ro)
		if err != nil {
			body = bytes.ReplaceAll(
				body,
				[]byte(`"extendedArtists":false`),
				[]byte(`"extendedArtists":{}`))
		}
	case *TopTenTorrentsResponse: // hack around orpheus bug in top 10
		err := json.Unmarshal(body, ro)
		if err != nil {
			body = bytes.ReplaceAll(
				body,
				[]byte(`"artist":false`),
				[]byte(`"artist":""`))
		}
	default:
	}
	return json.Unmarshal(body, responseObj)
}

type GenericResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func (w ClientStruct) Do(action string, params url.Values, result interface{}) error {
	requestURL, err := buildURL(w.baseURL, "ajax.php", action, params)
	if err != nil {
		return err
	}
	return w.GetJSON(requestURL, result)
}

//CreateDownloadURL constructs a download URL using the provided torrent id.
func (w ClientStruct) CreateDownloadURL(id int) (string, error) {
	if !w.loggedIn {
		return "", errRequestFailedLogin
	}

	params := url.Values{}
	params.Set("action", "download")
	params.Set("id", strconv.Itoa(id))
	params.Set("authkey", w.authkey)
	params.Set("torrent_pass", w.passkey)
	downloadURL, err := buildURL(w.baseURL, "torrents.php", "", params)
	if err != nil {
		return "", err
	}
	return downloadURL, nil
}

//CreateUploadURL constructs an upload URL for this tracker, and returns the
// url and autheky
func (w ClientStruct) CreateUploadURL() (u url.URL, a string, err error) {
	if !w.loggedIn {
		return u, a, errRequestFailedLogin
	}

	a = w.authkey
	u = w.baseURL
	u.Path = "upload.php"
	return u, a, err
}

func (w *ClientStruct) getCookies() error {
	if w.db == nil {
		return nil
	}
	var (
		c  []byte
		cs []*http.Cookie
	)
	err := w.db.QueryRow(`SELECT cookie FROM cookies WHERE url=?`,
		w.baseURL.String()).Scan(&c)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(c, &cs)
	if err == nil {
		w.client.Jar.SetCookies(&w.baseURL, cs)
	}
	return err
}

func (w *ClientStruct) clearCookies() (err error) {
	w.client.Jar, err = cookiejar.New(nil)
	if err != nil {
		return err
	}
	return w.saveCookies()
}

func (w *ClientStruct) saveCookies() error {
	// remember cookies
	if w.db == nil {
		return nil
	}
	cs := w.client.Jar.Cookies(&w.baseURL)
	c, err := json.Marshal(cs)
	if err != nil {
		return err
	}
	_, err = w.db.Exec(`REPLACE INTO cookies VALUES(?,?)`,
		w.baseURL.String(), c)
	return err
}

//Login logs in to the API using the provided credentials.
func (w *ClientStruct) Login(username, password string) error {
	if w.db != nil {
		err := w.getCookies() // sets cookie jar
		if err != nil {
			return err
		}
		// can get account without posting a login?
		err = w.GetAccount()
		if err == nil {
			w.loggedIn = true
			return nil
		}
		// nope, clear cookies and log in fresh
		err = w.clearCookies()
		if err != nil {
			return err
		}
	}
	params := url.Values{}
	params.Set("username", username)
	params.Set("password", password)

	reqBody := strings.NewReader(params.Encode())
	req, err := http.NewRequest("POST", w.baseURL.String()+"login.php", reqBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", w.userAgent)
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if !strings.Contains(resp.Request.URL.String(), "index") {
		return errLoginFailed
	}
	w.loggedIn = true
	err = w.GetAccount()
	if err != nil {
		return err
	}
	err = w.saveCookies()
	return err
}

//Logout logs out of the API, ending the current session.
func (w *ClientStruct) Logout() error {
	params := url.Values{"auth": {w.authkey}}
	requestURL, err := buildURL(w.baseURL, "logout.php", "", params)
	if err != nil {
		return err
	}
	_, err = w.client.Get(requestURL)
	if err != nil {
		return err
	}
	w.loggedIn, w.authkey, w.passkey = false, "", ""
	return nil
}

//GetAccount retrieves account information for the current user.
func (w *ClientStruct) GetAccount() error {
	account := AccountResponse{}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "index", url.Values{})
	if err != nil {
		return err
	}
	// don't cache login results
	db, loggedIn := w.db, w.loggedIn
	w.db, w.loggedIn = nil, true
	err = w.GetJSON(requestURL, &account)
	w.db, w.loggedIn = db, loggedIn
	if err != nil {
		return err
	}
	err = checkResponseStatus(account.Status, account.Error)
	if err != nil {
		return err
	}
	w.authkey, w.passkey = account.Response.AuthKey, account.Response.PassKey
	return nil
}

//GetMailbox retrieves mailbox information for the current user using the provided parameters.
func (w *ClientStruct) GetMailbox(params url.Values) (Mailbox, error) {
	mailbox := MailboxResponse{}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "inbox", params)
	if err != nil {
		return mailbox.Response, err
	}
	err = w.GetJSON(requestURL, &mailbox)
	if err != nil {
		return mailbox.Response, err
	}
	return mailbox.Response, checkResponseStatus(mailbox.Status, mailbox.Error)
}

//GetConversation retrieves conversation information for the current user using the provided conversation id and parameters.
func (w *ClientStruct) GetConversation(id int) (Conversation, error) {
	conversation := ConversationResponse{}
	params := url.Values{}
	params.Set("type", "viewconv")
	params.Set("id", strconv.Itoa(id))
	requestURL, err := buildURL(w.baseURL, "ajax.php", "inbox", params)
	if err != nil {
		return conversation.Response, err
	}
	err = w.GetJSON(requestURL, &conversation)
	if err != nil {
		return conversation.Response, err
	}
	return conversation.Response, checkResponseStatus(conversation.Status, conversation.Error)
}

//GetNotifications retrieves notification information using the specifed parameters.
func (w *ClientStruct) GetNotifications(params url.Values) (Notifications, error) {
	notifications := NotificationsResponse{}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "notifications", params)
	if err != nil {
		return notifications.Response, err
	}
	err = w.GetJSON(requestURL, &notifications)
	if err != nil {
		return notifications.Response, err
	}
	return notifications.Response, checkResponseStatus(notifications.Status, notifications.Error)
}

//GetAnnouncements retrieves announcement information.
func (w *ClientStruct) GetAnnouncements() (Announcements, error) {
	params := url.Values{}
	announcements := AnnouncementsResponse{}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "announcements", params)
	if err != nil {
		return announcements.Response, err
	}
	err = w.GetJSON(requestURL, &announcements)
	if err != nil {
		return announcements.Response, err
	}
	return announcements.Response, checkResponseStatus(announcements.Status, announcements.Error)
}

//GetSubscriptions retrieves forum subscription information for the current user using the provided parameters.
func (w *ClientStruct) GetSubscriptions(params url.Values) (Subscriptions, error) {
	subscriptions := SubscriptionsResponse{}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "subscriptions", params)
	if err != nil {
		return subscriptions.Response, err
	}
	err = w.GetJSON(requestURL, &subscriptions)
	if err != nil {
		return subscriptions.Response, err
	}
	return subscriptions.Response, checkResponseStatus(subscriptions.Status, subscriptions.Error)
}

//GetCategories retrieves forum category information.
func (w *ClientStruct) GetCategories() (Categories, error) {
	categories := CategoriesResponse{}
	params := url.Values{}
	params.Set("type", "main")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "forum", params)
	if err != nil {
		return categories.Response, err
	}
	err = w.GetJSON(requestURL, &categories)
	if err != nil {
		return categories.Response, err
	}
	return categories.Response, checkResponseStatus(categories.Status, categories.Error)
}

//GetForum retrieves forum information using the provided forum id and parameters.
func (w *ClientStruct) GetForum(id int, params url.Values) (Forum, error) {
	forum := ForumResponse{}
	params.Set("type", "viewforum")
	params.Set("forumid", strconv.Itoa(id))
	requestURL, err := buildURL(w.baseURL, "ajax.php", "forum", params)
	if err != nil {
		return forum.Response, err
	}
	err = w.GetJSON(requestURL, &forum)
	if err != nil {
		return forum.Response, err
	}
	return forum.Response, checkResponseStatus(forum.Status, forum.Error)
}

//GetThread retrieves forum thread information using the provided thread id and parameters.
func (w *ClientStruct) GetThread(id int, params url.Values) (Thread, error) {
	thread := ThreadResponse{}
	params.Set("type", "viewthread")
	params.Set("threadid", strconv.Itoa(id))
	requestURL, err := buildURL(w.baseURL, "ajax.php", "forum", params)
	if err != nil {
		return thread.Response, err
	}
	err = w.GetJSON(requestURL, &thread)
	if err != nil {
		return thread.Response, err
	}
	return thread.Response, checkResponseStatus(thread.Status, thread.Error)
}

//GetArtistBookmarks retrieves artist bookmark information for the current user.
func (w *ClientStruct) GetArtistBookmarks() (ArtistBookmarks, error) {
	artistBookmarks := ArtistBookmarksResponse{}
	params := url.Values{}
	params.Set("type", "artists")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "bookmarks", params)
	if err != nil {
		return artistBookmarks.Response, err
	}
	err = w.GetJSON(requestURL, &artistBookmarks)
	if err != nil {
		return artistBookmarks.Response, err
	}
	return artistBookmarks.Response, checkResponseStatus(artistBookmarks.Status, artistBookmarks.Error)
}

//GetTorrentBookmarks retrieves torrent bookmark information for the current user.
func (w *ClientStruct) GetTorrentBookmarks() (TorrentBookmarks, error) {
	torrentBookmarks := TorrentBookmarksResponse{}
	params := url.Values{}
	params.Set("type", "torrents")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "bookmarks", params)
	if err != nil {
		return torrentBookmarks.Response, err
	}
	err = w.GetJSON(requestURL, &torrentBookmarks)
	if err != nil {
		return torrentBookmarks.Response, err
	}
	return torrentBookmarks.Response, checkResponseStatus(torrentBookmarks.Status, torrentBookmarks.Error)
}

//GetArtist retrieves artist information using the provided artist id and parameters.
func (w *ClientStruct) GetArtist(id int, params url.Values) (Artist, error) {
	artist := ArtistResponse{}
	if _, ok := params["artistname"]; !ok || id != 0 {
		params.Set("id", strconv.Itoa(id))
	}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "artist", params)
	if err != nil {
		return artist.Response, err
	}
	err = w.GetJSON(requestURL, &artist)
	if err != nil {
		return artist.Response, err
	}
	return artist.Response, checkResponseStatus(artist.Status, artist.Error)
}

//GetRequest retrieves request information using the provided request id and parameters.
func (w *ClientStruct) GetRequest(id int, params url.Values) (Request, error) {
	request := RequestResponse{}
	params.Set("id", strconv.Itoa(id))
	requestURL, err := buildURL(w.baseURL, "ajax.php", "request", params)
	if err != nil {
		return request.Response, err
	}
	err = w.GetJSON(requestURL, &request)
	if err != nil {
		return request.Response, err
	}
	return request.Response, checkResponseStatus(request.Status, request.Error)
}

//GetTorrent retrieves torrent information using the provided torrent id and parameters.
func (w *ClientStruct) GetTorrent(id int, params url.Values) (GetTorrentStruct, error) {
	torrent := TorrentResponse{}
	if _, ok := params["hash"]; !ok || id != 0 {
		params.Set("id", strconv.Itoa(id))
	}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "torrent", params)
	if err != nil {
		return torrent.Response, err
	}
	err = w.GetJSON(requestURL, &torrent)
	if err != nil {
		return torrent.Response, err
	}
	return torrent.Response, checkResponseStatus(torrent.Status, torrent.Error)
}

//GetTorrentGroup retrieves torrent group information using the provided torrent group id and parameters.
func (w *ClientStruct) GetTorrentGroup(id int, params url.Values) (TorrentGroup, error) {
	torrentGroup := TorrentGroupResponse{}
	if _, ok := params["hash"]; !ok || id != 0 {
		params.Set("id", strconv.Itoa(id))
	}
	requestURL, err := buildURL(w.baseURL, "ajax.php", "torrentgroup", params)
	if err != nil {
		return torrentGroup.Response, err
	}
	err = w.GetJSON(requestURL, &torrentGroup)
	if err != nil {
		return torrentGroup.Response, err
	}
	return torrentGroup.Response, checkResponseStatus(torrentGroup.Status, torrentGroup.Error)
}

//SearchTorrents retrieves torrent search results using the provided search string and parameters.
func (w *ClientStruct) SearchTorrents(searchStr string, params url.Values) (TorrentSearch, error) {
	torrentSearch := TorrentSearchResponse{}
	params.Set("searchstr", searchStr)
	requestURL, err := buildURL(w.baseURL, "ajax.php", "browse", params)
	if err != nil {
		return torrentSearch.Response, err
	}
	err = w.GetJSON(requestURL, &torrentSearch)
	if err != nil {
		return torrentSearch.Response, err
	}
	return torrentSearch.Response, checkResponseStatus(torrentSearch.Status, torrentSearch.Error)
}

//SearchRequests retrieves request search results using the provided search string and parameters.
func (w *ClientStruct) SearchRequests(searchStr string, params url.Values) (RequestsSearch, error) {
	requestsSearch := RequestsSearchResponse{}
	params.Set("search", searchStr)
	requestURL, err := buildURL(w.baseURL, "ajax.php", "requests", params)
	if err != nil {
		return requestsSearch.Response, err
	}
	err = w.GetJSON(requestURL, &requestsSearch)
	if err != nil {
		return requestsSearch.Response, err
	}
	return requestsSearch.Response, checkResponseStatus(requestsSearch.Status, requestsSearch.Error)
}

//SearchUsers retrieves user search results using the provided search string and parameters.
func (w *ClientStruct) SearchUsers(searchStr string, params url.Values) (UserSearch, error) {
	userSearch := UserSearchResponse{}
	params.Set("search", searchStr)
	requestURL, err := buildURL(w.baseURL, "ajax.php", "usersearch", params)
	if err != nil {
		return userSearch.Response, err
	}
	err = w.GetJSON(requestURL, &userSearch)
	if err != nil {
		return userSearch.Response, err
	}
	return userSearch.Response, checkResponseStatus(userSearch.Status, userSearch.Error)
}

//GetTopTenTorrents retrieves "top ten torrents" information using the provided parameters.
func (w *ClientStruct) GetTopTenTorrents(params url.Values) (TopTenTorrents, error) {
	topTenTorrents := TopTenTorrentsResponse{}
	params.Set("type", "torrents")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "top10", params)
	if err != nil {
		return topTenTorrents.Response, err
	}
	err = w.GetJSON(requestURL, &topTenTorrents)
	if err != nil {
		return topTenTorrents.Response, err
	}
	return topTenTorrents.Response, checkResponseStatus(topTenTorrents.Status, topTenTorrents.Error)
}

//GetTopTenTags retrieves "top ten tags" information using the provided parameters.
func (w *ClientStruct) GetTopTenTags(params url.Values) (TopTenTags, error) {
	topTenTags := TopTenTagsResponse{}
	params.Set("type", "tags")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "top10", params)
	if err != nil {
		return topTenTags.Response, err
	}
	err = w.GetJSON(requestURL, &topTenTags)
	if err != nil {
		return topTenTags.Response, err
	}
	return topTenTags.Response, checkResponseStatus(topTenTags.Status, topTenTags.Error)
}

//GetTopTenUsers retrieves "top tem users" information using the provided parameters.
func (w *ClientStruct) GetTopTenUsers(params url.Values) (TopTenUsers, error) {
	topTenUsers := TopTenUsersResponse{}
	params.Set("type", "users")
	requestURL, err := buildURL(w.baseURL, "ajax.php", "top10", params)
	if err != nil {
		return topTenUsers.Response, err
	}
	err = w.GetJSON(requestURL, &topTenUsers)
	if err != nil {
		return topTenUsers.Response, err
	}
	return topTenUsers.Response, checkResponseStatus(topTenUsers.Status, topTenUsers.Error)
}

//GetSimilarArtists retrieves similar artist information using the provided artist id and limit.
func (w *ClientStruct) GetSimilarArtists(id, limit int) (SimilarArtists, error) {
	similarArtists := SimilarArtists{}
	params := url.Values{}
	params.Set("id", strconv.Itoa(id))
	params.Set("limit", strconv.Itoa(limit))
	requestURL, err := buildURL(w.baseURL, "ajax.php", "similar_artists", params)
	if err != nil {
		return similarArtists, err
	}
	err = w.GetJSON(requestURL, &similarArtists)
	if err != nil {
		return similarArtists, err
	}
	return similarArtists, nil
}
