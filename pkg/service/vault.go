package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/AleksZelenchuk/vault-server/gen/go/vaultpb"
	"github.com/AleksZelenchuk/vault-server/pkg/auth"
	"github.com/AleksZelenchuk/vault-server/pkg/storage"
	"github.com/google/uuid"
	_ "log"
	"reflect"
)

type VaultService struct {
	vaultpb.UnimplementedVaultServiceServer
	store *storage.Store
	// publisher can be used for Redis PubSub broadcasting
}

func NewVaultService(store *storage.Store) *VaultService {
	return &VaultService{store: store}
}

// CreateEntry create entry from given data
// user_id value is taken from an active user and cannot be passed to avoid data consistency problems
func (s *VaultService) CreateEntry(ctx context.Context, req *vaultpb.CreateEntryRequest) (*vaultpb.CreateEntryResponse, error) {
	userId, errValidate := auth.UserIDFromContext(ctx)
	if errValidate != true {
		return nil, errors.New("no user id provided")
	}

	if !validateEntry(req) {
		return nil, errors.New("invalid entry data")
	}

	newUuid := uuid.New()
	entry := &storage.Entry{
		ID:       newUuid,
		UserId:   userId,
		Title:    req.Entry.Title,
		Username: req.Entry.Username,
		Password: []byte(req.Entry.Password),
		Notes:    sqlNull(req.Entry.Notes),
		Tags:     req.Entry.Tags,
		Folder:   sqlNull(req.Entry.Folder),
		Domain:   sqlNull(req.Entry.Domain),
	}
	result, err := s.store.Create(ctx, entry)
	if err != nil {
		return nil, err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &vaultpb.CreateEntryResponse{Id: newUuid.String()}, nil
}

// validateEntry need to validate entry data to make sure required fields are there to avoid panic
func validateEntry(req *vaultpb.CreateEntryRequest) bool {
	entry := req.Entry
	if entry == nil {
		return false
	}
	v := reflect.ValueOf(entry).Elem()
	for _, field := range []string{"Title", "Username", "Password"} {
		val := v.FieldByName(field)
		if !val.IsValid() || val.Kind() != reflect.String || val.String() == "" {
			return false
		}
	}

	// ToDo for future check this field
	/*tags := v.FieldByName("Tags")
	if !tags.IsValid() || tags.Kind() != reflect.Slice || tags.Len() == 0 {
		return false
	}*/

	return true
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

// ListEntries here we retrieve list of all entries eligible for active user
func (s *VaultService) ListEntries(ctx context.Context, req *vaultpb.ListEntriesRequest) (*vaultpb.ListEntriesResponse, error) {
	_, errValidate := auth.UserIDFromContext(ctx)
	if errValidate != true {
		return nil, errors.New("no user id provided")
	}

	resp, err := s.store.List(ctx, req.Domain, req.Folder, req.Tags)
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
		Domain:   e.Domain.String,
	}
}

func sqlNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
