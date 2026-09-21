package scraper

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const chromeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func siteKey(pageURL string) string {
	switch {
	case strings.Contains(pageURL, "3xchina.page"):
		return "3xchina"
	case strings.Contains(pageURL, "muskuduu.com"):
		return "muskuduu"
	}
	return ""
}

// ── Video listing ─────────────────────────────────────────────────────────────

type VideoCard struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	PageURL    string    `json:"-"`
	Thumbnail  string    `json:"thumbnail"`
	Site       string    `json:"site"`
	Actors     []string  `json:"-"`
	Tags       []string  `json:"-"`
	Categories []string  `json:"-"`
	SyncedAt   time.Time `json:"synced_at,omitempty"`
}

func ScrapeVideoList(siteURL string, page int) ([]VideoCard, error) {
	switch siteKey(siteURL) {
	case "3xchina":
		return scrape3xchinaList(page)
	case "muskuduu":
		return scrapeMuskuduuList(page)
	default:
		return nil, fmt.Errorf("unsupported domain")
	}
}

// ── 3xchina.page ──────────────────────────────────────────────────────────────
//
// 3xchina renders its first listing page with JavaScript (no articles in the raw
// HTML), but server-renders the homepage and every /category/<cat>/page/N/ for
// N >= 2. So: page 1 comes from the homepage (newest ~36), later pages from the
// "chinese-av" category which holds every post and paginates correctly.
func scrape3xchinaList(page int) ([]VideoCard, error) {
	listURL := "https://3xchina.page/"
	if page >= 2 {
		listURL = fmt.Sprintf("https://3xchina.page/category/chinese-av/page/%d/", page)
	}

	doc, _, err := fetchDoc(listURL)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var cards []VideoCard
	doc.Find("article.loop-video, article[data-post-id]").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a[href]").First()
		link, _ := a.Attr("href")
		link = strings.TrimSpace(link)
		if link == "" || seen[link] || !strings.Contains(link, "3xchina.page/") {
			return
		}
		if strings.Contains(link, "/category/") || strings.Contains(link, "/feed/") ||
			strings.Contains(link, "/wp-json") || strings.TrimRight(link, "/") == "https://3xchina.page" {
			return
		}

		title, _ := a.Attr("title")
		if strings.TrimSpace(title) == "" {
			title = s.Find(".entry-header, h2, h3").First().Text()
		}

		thumb := ""
		img := s.Find("img").First()
		for _, attr := range []string{"data-src", "data-lazy-src", "data-original", "src"} {
			if v, ok := img.Attr(attr); ok && v != "" && !strings.HasPrefix(v, "data:") {
				thumb = v
				break
			}
		}

		// The <article> class list carries the taxonomy:
		//   actors-<slug> · tag-<slug> · category-<slug>
		var actors, tags, cats []string
		if cl, ok := s.Attr("class"); ok {
			for _, cls := range strings.Fields(cl) {
				switch {
				case strings.HasPrefix(cls, "actors-"):
					actors = append(actors, strings.TrimPrefix(cls, "actors-"))
				case strings.HasPrefix(cls, "tag-"):
					tags = append(tags, strings.TrimPrefix(cls, "tag-"))
				case strings.HasPrefix(cls, "category-"):
					cats = append(cats, strings.TrimPrefix(cls, "category-"))
				}
			}
		}

		seen[link] = true
		cards = append(cards, VideoCard{
			Title:      html.UnescapeString(strings.TrimSpace(title)),
			PageURL:    link,
			Thumbnail:  thumb,
			Site:       "3xchina",
			Actors:     actors,
			Tags:       tags,
			Categories: cats,
		})
	})

	if len(cards) == 0 {
		return nil, fmt.Errorf("3xchina: no listings on %s", listURL)
	}
	return cards, nil
}

// ── muskuduu.com ──────────────────────────────────────────────────────────────
//
// A WordPress JAV/Myanmar-sub site. Every listing page (including the homepage)
// is server-rendered with a plain <article data-video-id data-main-thumb> per
// video, so no JS rendering is needed. Genre is a readable class slug
// (category-<slug>); its "tag-<id>" classes are raw WordPress term ids with no
// name attached, so they're not usable and are skipped.
var reNumeric = regexp.MustCompile(`^[0-9]+$`)

func scrapeMuskuduuList(page int) ([]VideoCard, error) {
	listURL := "https://muskuduu.com/"
	if page >= 2 {
		listURL = fmt.Sprintf("https://muskuduu.com/page/%d/", page)
	}

	doc, _, err := fetchDoc(listURL)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var cards []VideoCard
	doc.Find("article[data-video-id]").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a[href]").First()
		link, _ := a.Attr("href")
		link = strings.TrimSpace(link)
		if link == "" || seen[link] || !strings.Contains(link, "muskuduu.com/") {
			return
		}

		title, _ := a.Attr("title")
		if strings.TrimSpace(title) == "" {
			title = s.Find(".entry-header").First().Text()
		}

		thumb, _ := s.Attr("data-main-thumb")
		if thumb == "" {
			img := s.Find("img").First()
			for _, attr := range []string{"data-src", "data-lazy-src", "src"} {
				if v, ok := img.Attr(attr); ok && v != "" && !strings.HasPrefix(v, "data:") {
					thumb = v
					break
				}
			}
		}

		var cats []string
		if cl, ok := s.Attr("class"); ok {
			for _, cls := range strings.Fields(cl) {
				if slug := strings.TrimPrefix(cls, "category-"); slug != cls && !reNumeric.MatchString(slug) {
					cats = append(cats, slug)
				}
			}
		}

		seen[link] = true
		cards = append(cards, VideoCard{
			Title:      html.UnescapeString(strings.TrimSpace(title)),
			PageURL:    link,
			Thumbnail:  thumb,
			Site:       "muskuduu",
			Categories: cats,
		})
	})

	if len(cards) == 0 {
		return nil, fmt.Errorf("muskuduu: no listings on %s", listURL)
	}
	return cards, nil
}

// The video page embeds an iframe whose ?q= is base64 of a query string
// carrying a "tag" param — the URL-encoded HTML of a <video><source src=...>
// tag pointing at cdn.aiedit.top, which itself 302s to a signed, unlocked S3 URL.
var reMuskuduuPlayer = regexp.MustCompile(`player-x\.php\?q=([A-Za-z0-9+/=]+)`)

func scrapeMuskuduu(pageURL string) (*VideoResult, error) {
	_, rawHTML, err := fetchDoc(pageURL)
	if err != nil {
		return nil, err
	}

	m := reMuskuduuPlayer.FindStringSubmatch(rawHTML)
	if m == nil {
		return nil, fmt.Errorf("muskuduu: no player iframe on page")
	}
	decoded, err := base64.StdEncoding.DecodeString(m[1])
	if err != nil {
		return nil, fmt.Errorf("muskuduu: bad player payload: %w", err)
	}
	qs, err := url.ParseQuery(string(decoded))
	if err != nil {
		return nil, fmt.Errorf("muskuduu: bad player query: %w", err)
	}
	tag := qs.Get("tag")

	srcM := regexp.MustCompile(`<source[^>]*\ssrc="([^"]+)"`).FindStringSubmatch(tag)
	if srcM == nil {
		return nil, fmt.Errorf("muskuduu: no video source in player tag")
	}

	thumb := ""
	if posterM := regexp.MustCompile(`poster="([^"]+)"`).FindStringSubmatch(tag); posterM != nil {
		thumb = html.UnescapeString(posterM[1])
	}

	return &VideoResult{URL: html.UnescapeString(srcM[1]), Type: "mp4", Thumb: thumb}, nil
}

// ── Single video URL ──────────────────────────────────────────────────────────

type VideoResult struct {
	URL     string   `json:"url"`
	Type    string   `json:"type"`
	Mirrors []string `json:"mirrors,omitempty"`
	Thumb   string   `json:"-"` // poster image, when the host exposes one
}

func ScrapeVideoURL(pageURL string) (*VideoResult, error) {
	switch siteKey(pageURL) {
	case "3xchina":
		// The extraction is a live chain of third-party requests (3xchina →
		// hglink → audinifer); a single hiccup shouldn't surface as a playback failure.
		var err error
		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
			}
			var r *VideoResult
			if r, err = scrapeStreamHG(pageURL); err == nil {
				return r, nil
			}
		}
		return nil, err
	case "muskuduu":
		return scrapeMuskuduu(pageURL)
	default:
		return nil, fmt.Errorf("unsupported domain")
	}
}

// ── StreamHG / hglink.to  (3xchina.page) ─────────────────────────────────────
//
// 3xchina.page embeds videos through <iframe src="https://<front>/e/<id>">, where
// <front> rotates between StreamHG loader domains (hglink.to, gradehgplus.com, …).
// Every front is just a JS redirect to the same StreamHG backend, whose /e/<id>
// page carries a Dean-Edwards-packed player config with the real HLS sources.
// Everything below is plain HTTP — no headless browser.
//
// If StreamHG rotates its backend domain, update streamHGHost
// (it is whatever <front>/main.js sets `location.href` to).
const streamHGHost = "https://audinifer.com"

// matches the embed id in an <iframe src="https://<anything>/e/<id>"> (also /v/, /d/, /embed/).
var reEmbedID = regexp.MustCompile(`(?i)https?://[a-z0-9.-]+/(?:e|v|d|embed)/([a-z0-9]{8,16})`)

func scrapeStreamHG(pageURL string) (*VideoResult, error) {
	doc, rawHTML, err := fetchDoc(pageURL)
	if err != nil {
		return nil, err
	}

	// Locate the embed id — prefer an <iframe> inside the player, then anywhere.
	id := ""
	doc.Find(".video-player iframe[src], .responsive-player iframe[src], iframe[src]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if src, _ := s.Attr("src"); reEmbedID.MatchString(src) {
			id = reEmbedID.FindStringSubmatch(src)[1]
			return false
		}
		return true
	})
	if id == "" {
		if m := reEmbedID.FindStringSubmatch(rawHTML); m != nil {
			id = m[1]
		}
	}
	if id == "" {
		return nil, fmt.Errorf("streamhg: no embed iframe on page")
	}

	// Fetch the StreamHG backend embed page (a Referer is required, any value works).
	req, _ := http.NewRequest("GET", streamHGHost+"/e/"+id, nil)
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("Referer", "https://hglink.to/")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("streamhg: embed fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("streamhg: embed status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	js := unpackPacked(string(body))
	if js == "" {
		js = string(body) // not every build is packed
	}

	// The player config exposes up to three HLS sources, but only one is usable
	// from a browser: hls2 carries its own token and returns permissive CORS from
	// any origin. hls4 (audinifer /stream) sends no CORS header and hls3's domain
	// is frequently dead — offering them as "servers" would just be dead buttons.
	// So take the first available in priority order and stop.
	norm := func(u string) string {
		u = strings.NewReplacer(`\/`, `/`, `\\`, ``).Replace(strings.TrimSpace(u))
		if strings.HasPrefix(u, "/") {
			u = streamHGHost + u
		}
		return u
	}
	pick := ""
	for _, key := range []string{"hls2", "hls3", "hls4"} {
		if m := regexp.MustCompile(`"` + key + `"\s*:\s*"([^"]+)"`).FindStringSubmatch(js); m != nil {
			if u := norm(m[1]); strings.HasPrefix(u, "http") {
				pick = u
				break
			}
		}
	}
	if pick == "" {
		if m := regexp.MustCompile(`https?://[^\s"'\\]+/master\.(?:m3u8|txt)[^\s"'\\]*`).FindString(js); m != "" {
			pick = norm(m)
		}
	}
	if pick == "" {
		return nil, fmt.Errorf("streamhg: no video source in embed")
	}
	return &VideoResult{URL: pick, Type: "hls"}, nil
}

// unpackPacked reverses a Dean Edwards p.a.c.k.e.r payload of the form
//
//	eval(function(p,a,c,k,e,d){…}('<payload>',<radix>,<count>,'<w0>|<w1>|…'.split('|'),0,{}))
//
// Returns "" when src is not packed.
func unpackPacked(src string) string {
	m := regexp.MustCompile(`\}\(\s*'((?:[^'\\]|\\.)*)'\s*,\s*(\d+)\s*,\s*(\d+)\s*,\s*'([^']*)'\.split\('\|'\)`).FindStringSubmatch(src)
	if m == nil {
		return ""
	}
	payload := strings.NewReplacer(`\\`, `\`, `\'`, `'`, `\"`, `"`).Replace(m[1])
	radix, _ := strconv.Atoi(m[2])
	count, _ := strconv.Atoi(m[3])
	words := strings.Split(m[4], "|")
	if radix < 2 {
		radix = 36
	}

	dict := make(map[string]string, count)
	for i := 0; i < count; i++ {
		key := encodeBase(i, radix)
		if i < len(words) && words[i] != "" {
			dict[key] = words[i]
		} else {
			dict[key] = key
		}
	}
	return regexp.MustCompile(`\b\w+\b`).ReplaceAllStringFunc(payload, func(tok string) string {
		if v, ok := dict[tok]; ok {
			return v
		}
		return tok
	})
}

// ── Generic video-host embed (StreamWish / VidHide / StreamHG family) ─────────
//
// These XFileShare-style hosts all serve an /e/<code> embed page carrying a
// jwplayer config (often Dean-Edwards-packed) with an HLS master playlist.
// ScrapeEmbed pulls that m3u8 out. Used for manually-added "Myanmar" videos.

// EmbedHostRe recognises the hosts ScrapeEmbed knows how to unpack. Anything
// else that looks like an embed page is played through an <iframe> instead.
var EmbedHostRe = regexp.MustCompile(`(?i)\b(streamwish|embedwish|wishembed|swishsrv|awish|dwish|hlswish|streamhg|vidhide|streamhide|filelions|vidhidevip|hglink|gradehgplus|audinifer|filemoon|kerapoxy)\b`)

var embedSourceRes = []*regexp.Regexp{
	regexp.MustCompile(`"hls2"\s*:\s*"([^"]+)"`), // StreamHG family: hls2 has best CORS
	regexp.MustCompile(`sources?\s*[:=]\s*\[\s*\{[^}]*?file\s*:\s*["']([^"']+\.m3u8[^"']*)["']`),
	regexp.MustCompile(`["']?file["']?\s*:\s*["']([^"']+\.m3u8[^"']*)["']`),
	regexp.MustCompile(`["'](https?://[^\s"'\\]+/master\.(?:m3u8|txt)[^\s"'\\]*)["']`),
	regexp.MustCompile(`(https?://[^\s"'\\<>]+\.m3u8[^\s"'\\<>]*)`),
}

// blocked video-host domains that 403 direct/iframe access, mapped to a
// sibling domain that still serves the /e/<code> player.
var embedDomainRemap = map[string]string{
	"streamwish.com": "embedwish.com",
	"streamwish.to":  "embedwish.com",
	"www.streamwish.com": "embedwish.com",
}

// NormalizeEmbedURL rewrites /f/ and /d/ paths to /e/ and swaps blocked hosts.
func NormalizeEmbedURL(embedURL string) string {
	embedURL = strings.Replace(strings.Replace(embedURL, "/f/", "/e/", 1), "/d/", "/e/", 1)
	if u, err := url.Parse(embedURL); err == nil {
		if alt, ok := embedDomainRemap[strings.ToLower(u.Host)]; ok {
			u.Host = alt
			embedURL = u.String()
		}
	}
	return embedURL
}

// ScrapeEmbed fetches an /e/<code> embed page and returns its HLS URL.
func ScrapeEmbed(embedURL string) (*VideoResult, error) {
	embedURL = NormalizeEmbedURL(embedURL)

	origin := embedURL
	if u, err := url.Parse(embedURL); err == nil {
		origin = u.Scheme + "://" + u.Host
	}

	req, _ := http.NewRequest("GET", embedURL, nil)
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("Referer", origin+"/")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	// Search the raw HTML and every unpacked packer payload it contains.
	haystacks := []string{string(body)}
	if up := unpackPacked(string(body)); up != "" {
		haystacks = append(haystacks, up)
	}

	thumb := ""
	reImage := regexp.MustCompile(`["'](https?://[^\s"'\\]+\.(?:jpg|jpeg|png|webp)[^\s"'\\]*)["']`)
	for _, hs := range haystacks {
		if thumb == "" {
			if m := reImage.FindStringSubmatch(hs); m != nil {
				thumb = strings.NewReplacer(`\/`, `/`, `\\`, ``).Replace(m[1])
			}
		}
		for _, re := range embedSourceRes {
			if m := re.FindStringSubmatch(hs); m != nil {
				u := strings.NewReplacer(`\/`, `/`, `\\`, ``).Replace(strings.TrimSpace(m[1]))
				if strings.HasPrefix(u, "http") {
					return &VideoResult{URL: u, Type: "hls", Thumb: thumb}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no m3u8 in embed page")
}

// encodeBase renders n the way the packer's `e` helper does:
// 0-9a-z for digit values < 36, then A-Z for 36-61.
func encodeBase(n, radix int) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		d := n % radix
		if d < 36 {
			out = string(digits[d]) + out
		} else {
			out = string(rune(d+29)) + out
		}
		n /= radix
	}
	return out
}

// ── HTTP helper ───────────────────────────────────────────────────────────────

func fetchDoc(pageURL string) (*goquery.Document, string, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", pageURL)

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse HTML: %w", err)
	}
	rawHTML, _ := doc.Html()
	return doc, rawHTML, nil
}
