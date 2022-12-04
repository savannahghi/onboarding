package fb

import (
	"context"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

// FirestoreClientExtension represents the methods we need from firebase `firestore.Client`
type FirestoreClientExtension interface {
	GetAll(ctx context.Context, query *GetAllQuery) ([]*firestore.DocumentSnapshot, error)
	Create(ctx context.Context, command *CreateCommand) (*firestore.DocumentRef, error)
	Update(ctx context.Context, command *UpdateCommand) error
	Delete(ctx context.Context, command *DeleteCommand) error
	Get(ctx context.Context, query *GetSingleQuery) (*firestore.DocumentSnapshot, error)
	RawClient(ctx context.Context) *firestore.Client
}

// FirestoreClientExtensionImpl ...
type FirestoreClientExtensionImpl struct {
	client *firestore.Client
}

// NewFirestoreClientExtension initializes a new FirestoreClient extension
func NewFirestoreClientExtension(fc *firestore.Client) *FirestoreClientExtensionImpl {
	return &FirestoreClientExtensionImpl{client: fc}
}

// GetAllQuery represent payload required to perform a request in the database
type GetAllQuery struct {
	CollectionName string
	FieldName      string
	Value          interface{}
	Operator       string
}

// GetSingleQuery represent payload required to get a single item from the database
type GetSingleQuery struct {
	CollectionName string
	Value          string
}

// CreateCommand represent payload required to perform a create operation in the database
type CreateCommand struct {
	CollectionName string
	Data           interface{}
}

// UpdateCommand represent payload required to perform an update operaion in the database
type UpdateCommand struct {
	CollectionName string
	Data           interface{}
	ID             string
}

// DeleteCommand represent payload required to perform a delete operation in the database
type DeleteCommand struct {
	CollectionName string
	ID             string
}

// GetAll retrieve a value from the store
func (f *FirestoreClientExtensionImpl) GetAll(ctx context.Context, getQuery *GetAllQuery) ([]*firestore.DocumentSnapshot, error) {
	collection := f.client.Collection(getQuery.CollectionName)

	var documents []*firestore.DocumentSnapshot

	if getQuery.FieldName == "" && getQuery.Operator == "" && getQuery.Value == nil {
		docs, err := collection.Documents(ctx).GetAll()
		if err != nil {
			return nil, exceptions.InternalServerError(err)
		}

		documents = docs

	} else {
		query := collection.Where(getQuery.FieldName, getQuery.Operator, getQuery.Value)
		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			return nil, exceptions.InternalServerError(err)
		}

		documents = docs
	}

	return documents, nil
}

// Create persists data to a firestore collection
func (f *FirestoreClientExtensionImpl) Create(ctx context.Context, command *CreateCommand) (*firestore.DocumentRef, error) {
	docRef, _, err := f.client.Collection(command.CollectionName).Add(ctx, command.Data)
	if err != nil {
		return nil, exceptions.InternalServerError(fmt.Errorf("unable to create new user profile: %w", err))
	}
	return docRef, nil
}

// Update updates data to a firestore collection
func (f *FirestoreClientExtensionImpl) Update(ctx context.Context, command *UpdateCommand) error {
	_, err := f.client.Collection(command.CollectionName).Doc(command.ID).Set(ctx, command.Data)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes data to a firestore collection
func (f *FirestoreClientExtensionImpl) Delete(ctx context.Context, command *DeleteCommand) error {
	_, err := f.client.Collection(command.CollectionName).Doc(command.ID).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves data to a firestore collection
func (f *FirestoreClientExtensionImpl) Get(ctx context.Context, query *GetSingleQuery) (*firestore.DocumentSnapshot, error) {
	dsnap, err := f.client.Collection(query.CollectionName).Doc(query.Value).Get(ctx)
	if err != nil {
		return nil, exceptions.InternalServerError(fmt.Errorf("unable to retrieve newly created user profile: %w", err))
	}
	return dsnap, nil
}

// RawClient returns the actual FirestoreClient
func (f *FirestoreClientExtensionImpl) RawClient(ctx context.Context) *firestore.Client {
	return f.client
}

// FirebaseClientExtension represents the methods we need from firebase `auth.Client`
type FirebaseClientExtension interface {
	GetUserByPhoneNumber(ctx context.Context, phone string) (*auth.UserRecord, error)
	CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error)
	DeleteUser(ctx context.Context, uid string) error
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseClientExtensionImpl ...
type FirebaseClientExtensionImpl struct{}

// NewFirebaseClientExtensionImpl initializes a new FirebaseClient extension
func NewFirebaseClientExtensionImpl() FirebaseClientExtension {
	return &FirebaseClientExtensionImpl{}
}

// GetUserByPhoneNumber ...
func (f *FirebaseClientExtensionImpl) GetUserByPhoneNumber(ctx context.Context, phone string) (*auth.UserRecord, error) {
	var client *auth.Client
	return client.GetUserByPhoneNumber(ctx, phone)
}

// CreateUser ...
func (f *FirebaseClientExtensionImpl) CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error) {
	var client *auth.Client
	return client.CreateUser(ctx, user)
}

// DeleteUser ...
func (f *FirebaseClientExtensionImpl) DeleteUser(ctx context.Context, uid string) error {
	var client *auth.Client
	return client.DeleteUser(ctx, uid)
}

// VerifyIDToken checks the validity of the provided id token and returns the uid
func (f *FirebaseClientExtensionImpl) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	var client *auth.Client
	authToken, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, exceptions.InternalServerError(fmt.Errorf("unable to verify the provided idToken: %w", err))
	}
	return authToken, nil
}
