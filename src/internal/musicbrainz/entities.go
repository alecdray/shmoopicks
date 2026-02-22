package musicbrainz

type EntityType string

var (
	EntityArea            EntityType = "area"
	EntityArtist          EntityType = "artist"
	EntityCollection      EntityType = "collection"
	EntityUserCollections EntityType = "user-collections"
	EntityEvent           EntityType = "event"
	EntityGenre           EntityType = "genre"
	EntityInstrument      EntityType = "instrument"
	EntityLabel           EntityType = "label"
	EntityPlace           EntityType = "place"
	EntityRecording       EntityType = "recording"
	EntityRelease         EntityType = "release"
	EntityReleaseGroup    EntityType = "release-group"
	EntitySeries          EntityType = "series"
	EntityWork            EntityType = "work"
	EntityURL             EntityType = "url"
)

func (e EntityType) String() string {
	return string(e)
}

func (e EntityType) SubEntities() []EntityType {
	switch e {
	case EntityArtist:
		return []EntityType{EntityRecording, EntityRelease, EntityReleaseGroup, EntityWork}
	case EntityCollection:
		return []EntityType{EntityUserCollections}
	case EntityLabel:
		return []EntityType{EntityRelease}
	case EntityRecording:
		return []EntityType{EntityRelease, EntityReleaseGroup}
	case EntityRelease:
		return []EntityType{EntityCollection, EntityLabel, EntityRecording, EntityReleaseGroup}
	case EntityReleaseGroup:
		return []EntityType{EntityRelease}
	default:
		return nil
	}
}

type Entity interface {
	Slug() EntityType
}

type Recording struct {
	ID               string         `json:"id"`
	Score            int            `json:"score"`
	ArtistCreditID   string         `json:"artist-credit-id"`
	Title            string         `json:"title"`
	Length           *int           `json:"length"`
	Video            *bool          `json:"video"`
	FirstReleaseDate string         `json:"first-release-date"`
	ArtistCredit     []ArtistCredit `json:"artist-credit"`
	Releases         []Release      `json:"releases"`
	ISRCs            []string       `json:"isrcs,omitempty"`
	Tags             []Tag          `json:"tags,omitempty"`
}

func (r Recording) Slug() EntityType {
	return EntityRecording
}

type ReleaseGroup struct {
	ID               string         `json:"id"`
	TypeID           string         `json:"type-id"`
	Score            int            `json:"score"`
	PrimaryTypeID    string         `json:"primary-type-id"`
	ArtistCreditID   string         `json:"artist-credit-id"`
	Count            int            `json:"count"`
	Title            string         `json:"title"`
	FirstReleaseDate string         `json:"first-release-date"`
	PrimaryType      string         `json:"primary-type"`
	ArtistCredit     []ArtistCredit `json:"artist-credit"`
	Releases         []Release      `json:"releases"`
	Tags             []Tag          `json:"tags"`
}

func (r ReleaseGroup) Slug() EntityType {
	return EntityReleaseGroup
}

type ReleaseEvent struct {
	Date string `json:"date"`
	Area Area   `json:"area"`
}

type Area struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	SortName      string   `json:"sort-name"`
	ISO31661Codes []string `json:"iso-3166-1-codes"`
}

func (r Area) Slug() EntityType {
	return EntityArea
}

type Media struct {
	ID          string  `json:"id"`
	Position    int     `json:"position"`
	Format      string  `json:"format,omitempty"`
	Track       []Track `json:"track"`
	TrackCount  int     `json:"track-count"`
	TrackOffset int     `json:"track-offset"`
}

type Track struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Title  string `json:"title"`
	Length *int   `json:"length,omitempty"`
}

type ArtistCredit struct {
	Name   string `json:"name"`
	Artist Artist `json:"artist"`
}

type Artist struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	SortName string  `json:"sort-name"`
	Aliases  []Alias `json:"aliases"`
}

func (r Artist) Slug() EntityType {
	return EntityArtist
}

type Alias struct {
	SortName  string  `json:"sort-name"`
	TypeID    string  `json:"type-id,omitempty"`
	Name      string  `json:"name"`
	Locale    *string `json:"locale"`
	Type      *string `json:"type"`
	Primary   *string `json:"primary"`
	BeginDate *string `json:"begin-date"`
	EndDate   *string `json:"end-date"`
}

type Release struct {
	ID             string         `json:"id"`
	StatusID       string         `json:"status-id,omitempty"`
	ArtistCreditID string         `json:"artist-credit-id"`
	Count          int            `json:"count"`
	Title          string         `json:"title"`
	Status         string         `json:"status,omitempty"`
	ArtistCredit   []ArtistCredit `json:"artist-credit"`
	ReleaseGroup   ReleaseGroup   `json:"release-group"`
	Date           string         `json:"date,omitempty"`
	Country        string         `json:"country,omitempty"`
	ReleaseEvents  []ReleaseEvent `json:"release-events,omitempty"`
	TrackCount     int            `json:"track-count"`
	Media          []Media        `json:"media"`
}

func (r Release) Slug() EntityType {
	return EntityRelease
}

type Tag struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}
