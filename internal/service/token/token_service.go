package token

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	tokenquery "github.com/fivemanage/lite/internal/database/query/token"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/fivemanage/lite/internal/service/token")

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db: db,
	}
}

func (r *Service) GetToken(ctx context.Context, apiToken string) (*database.Token, error) {
	// we need to add this check to the env check in cmd/lite/lite.go
	hmacSecret := os.Getenv("API_TOKEN_HMAC_SECRET")
	tokenHash := crypt.ComputeHMAC(hmacSecret, apiToken)

	token, err := tokenquery.SelectByHash(ctx, r.db, tokenHash)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *Service) CreateToken(ctx context.Context, data *api.CreateTokenRequest, userID int64) (string, error) {
	apiToken, err := crypt.GenerateApiKey()
	if err != nil {
		return "", err
	}

	hmacSecret := os.Getenv("API_TOKEN_HMAC_SECRET")
	tokenHash := crypt.ComputeHMAC(hmacSecret, apiToken)

	token := &database.Token{
		OrganizationID: data.OrganizationID,
		Identifier:     data.Identifier,
		TokenHash:      tokenHash,
		UserID:         int(userID),
	}

	err = tokenquery.Create(ctx, r.db, token)
	if err != nil {
		return "", err
	}

	return apiToken, nil
}

func (r *Service) ListTokens(ctx context.Context, organizationID string) ([]*api.ListTokensResponse, error) {
	var err error
	var response []*api.ListTokensResponse

	ctx, span := tracer.Start(ctx, "token.list_tokens")
	defer span.End()

	tokens, err := tokenquery.List(ctx, r.db, organizationID)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to list tokens", "err", err)
		return nil, err
	}

	for _, token := range tokens {
		response = append(response, &api.ListTokensResponse{
			ID:         token.ID,
			Identifier: token.Identifier,
		})
	}

	span.SetAttributes(
		attribute.Int("token_count", len(tokens)),
	)

	slog.Debug("listed tokens successfully", "count", len(tokens))

	return response, nil
}

func (r *Service) DeleteToken(ctx context.Context, organizationID, tokenID string) error {
	var err error
	iTokenID, err := strconv.ParseInt(tokenID, 10, 64)
	if err != nil {
		return err
	}

	err = tokenquery.Delete(ctx, r.db, organizationID, iTokenID)
	if err != nil {
		return err
	}
	return nil
}
