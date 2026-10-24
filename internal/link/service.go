package link

import (
	"context"
	"crypto/rand"
	"errors"
	"net/url"
)

var ErrInvalidURL = errors.New("invalid URL")
var ErrUnsupportedScheme = errors.New("unsupported URL scheme")
var ErrMissingHostname = errors.New("URL must contain a hostname")
var ErrCodeAttemptsExhausted = errors.New("could not generate a unique code")

const codeGenerateAttempts = 5

type Service struct {
	repository   Repository
	cache        Cache
	generateCode func() string
}

func NewService(
	repository Repository,
	cache Cache,
) *Service {
	return &Service{
		repository:   repository,
		cache:        cache,
		generateCode: rand.Text,
	}
}

func (s *Service) Create(ctx context.Context, originalURL string) (Link, error) {
	parsed, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return Link{}, ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Link{}, ErrUnsupportedScheme
	}
	if parsed.Hostname() == "" {
		return Link{}, ErrMissingHostname
	}

	for range codeGenerateAttempts {
		created := Link{Code: s.generateCode(), OriginalURL: originalURL}
		err := s.repository.Save(ctx, created)
		if err == nil {
			return created, nil
		}
		if !errors.Is(err, ErrCodeExists) {
			return Link{}, err
		}
	}
	return Link{}, ErrCodeAttemptsExhausted
}

func (s *Service) Get(ctx context.Context, code string) (Link, error) {
	if s.cache == nil {
		return s.repository.Get(ctx, code)
	}

	rValue, rBool, rErr := s.cache.Get(ctx, code)

	if rBool && rErr == nil {
		return Link{
			Code:        code,
			OriginalURL: rValue,
		}, nil
	}

	dbLink, err := s.repository.Get(ctx, code)

	if err == nil {
		rErr = s.cache.Set(
			ctx,
			code,
			dbLink.OriginalURL,
		)
	}

	return dbLink, err
}
