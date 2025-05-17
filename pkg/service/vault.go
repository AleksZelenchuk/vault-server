package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AleksZelenchuk/vault-server/gen/go/vaultpb"
	"github.com/AleksZelenchuk/vault-server/pkg/auth"
	_ "log"
	"strconv"

	"github.com/AleksZelenchuk/vault-server/pkg/storage"
	"github.com/google/uuid"
)

type VaultService struct {
	vaultpb.UnimplementedVaultServiceServer
	store *storage.Store
	// publisher can be used for Redis PubSub broadcasting
}

func NewVaultService(store *storage.Store) *VaultService {
	return &VaultService{store: store}
}

func (s *VaultService) CreateEntry(ctx context.Context, req *vaultpb.CreateEntryRequest) (*vaultpb.CreateEntryResponse, error) {
	_, errValidate := auth.UserIDFromContext(ctx)
	if errValidate != true {
		return nil, errors.New("no user id provided")
	}

	entry := &storage.Entry{
		ID:       uuid.New(),
		UserId:   req.Entry.UserId,
		Title:    req.Entry.Title,
		Username: req.Entry.Username,
		Password: []byte(req.Entry.Password),
		Notes:    sqlNull(req.Entry.Notes),
		Tags:     req.Entry.Tags,
		Folder:   sqlNull(req.Entry.Folder),
	}
	result, err := s.store.Create(ctx, entry)
	fmt.Println(result)
	if err != nil {
		return nil, err
	}
	lastInsertedId, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &vaultpb.CreateEntryResponse{Id: strconv.FormatInt(lastInsertedId, 10)}, nil
}

func (s *VaultService) GetEntry(ctx context.Context, req *vaultpb.GetEntryRequest) (*vaultpb.GetEntryResponse, error) {
	_, err := auth.UserIDFromContext(ctx)
	if err != true {
		return nil, errors.New("no user id provided")
	}

	id, err2 := uuid.Parse(req.Id)
	if err2 != nil {
		return nil, err2
	}

	entry, err2 := s.store.Get(ctx, id)
	if err2 != nil {
		return nil, err2
	}

	return &vaultpb.GetEntryResponse{Entry: toProto(entry)}, nil
}

func (s *VaultService) DeleteEntry(ctx context.Context, req *vaultpb.DeleteEntryRequest) (*vaultpb.DeleteEntryResponse, error) {
	_, errValidate := auth.UserIDFromContext(ctx)
	if errValidate != true {
		return nil, errors.New("no user id provided")
	}

	id, err2 := uuid.Parse(req.Id)
	if err2 != nil {
		return nil, err2
	}

	success, err2 := s.store.Delete(ctx, id)
	if err2 != nil {
		if errors.Is(err2, sql.ErrNoRows) {
			return nil, errors.New("entry not found")
		}
		return nil, err2
	}

	return &vaultpb.DeleteEntryResponse{Success: success}, nil
}

func (s *VaultService) ListEntries(ctx context.Context, req *vaultpb.ListEntriesRequest) (*vaultpb.ListEntriesResponse, error) {
	_, errValidate := auth.UserIDFromContext(ctx)
	if errValidate != true {
		return nil, errors.New("no user id provided")
	}

	resp, err := s.store.List(ctx, req.Folder, req.Tags)
	if err != nil {
		return nil, err
	}
	var vaultEntries []*vaultpb.VaultEntry
	for _, entry := range resp {
		entry.Password, _ = storage.Decrypt(entry.Password)
		vaultEntries = append(vaultEntries, toProto(&entry))
	}
	return &vaultpb.ListEntriesResponse{Entries: vaultEntries}, nil
}

// Helpers
func toProto(e *storage.Entry) *vaultpb.VaultEntry {
	return &vaultpb.VaultEntry{
		Id:       e.ID.String(),
		Title:    e.Title,
		Username: e.Username,
		Password: string(e.Password),
		Notes:    e.Notes.String,
		Tags:     e.Tags,
		Folder:   e.Folder.String,
	}
}

func sqlNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
