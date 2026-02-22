package musicbrainz

import (
	"fmt"
	"shmoopicks/src/internal/core/contextx"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Client() *Client {
	return s.client
}

func (s *Service) FindRecording(ctx contextx.ContextX, entity Entity, title string, artist string) (*Recording, error) {

	results, err := s.client.SearchEntities(ctx, Recording{}, QueryProps{
		Query: fmt.Sprintf("name: %s AND text: %s", title, artist),
		Limit: 5,
	})
	if err != nil {
		err = fmt.Errorf("failed to search musicbrainz: %w", err)
		return nil, err
	}

	for _, recording := range results.Recordings {
		isTitleMatch := fuzzy.RankMatchNormalizedFold(title, recording.Title) != -1
		if isTitleMatch {
			for _, recordingArtist := range recording.ArtistCredit {
				isArtistMatch := fuzzy.RankMatchNormalizedFold(artist, recordingArtist.Name) != -1
				if isArtistMatch {
					return &recording, nil
				}
			}
		}
	}

	return nil, nil
}
