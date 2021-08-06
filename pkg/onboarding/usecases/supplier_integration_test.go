package usecases_test

import (
	"context"
	"testing"

	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
	"github.com/stretchr/testify/assert"
)

const (
	testChargeMasterParentOrgId = "83d3479d-e902-4aab-a27d-6d5067454daf"
	testChargeMasterBranchID    = "94294577-6b27-4091-9802-1ce0f2ce4153"
	primaryEmail                = "test@bewell.co.ke"
)

func TestFindSupplierByUID(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}
	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service")
		return
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.Supplier
		wantErr bool
	}{
		{
			name: "happy :) find supplier by UID",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "sad :( fail to find supplier by UID",
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supplier, err := s.Supplier.FindSupplierByUID(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SupplierUseCasesImpl.FindSupplierByUID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if supplier != nil {
				if supplier.ID == "" {
					t.Errorf("expected a supplier.")
					return
				}
			}
		})
	}
}

func TestFindSupplierByID(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}
	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service")
		return
	}

	name := "Makmende And Sons"
	partnerPractitioner := profileutils.PartnerTypePractitioner
	partnerType, err := s.Supplier.AddPartnerType(ctx, &name, &partnerPractitioner)
	assert.Nil(t, err)
	assert.NotNil(t, partnerType)
	assert.Equal(t, true, partnerType)

	supplier, err := s.Supplier.SetUpSupplier(ctx, profileutils.AccountTypeOrganisation)
	assert.Nil(t, err)

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.Supplier
		wantErr bool
	}{
		{
			name: "happy :) find supplier by ID",
			args: args{
				ctx: ctx,
				id:  supplier.ID,
			},
			wantErr: false,
		},
		{
			name: "happy :) find supplier by ID using wrong context, should not fail",
			args: args{
				ctx: context.Background(),
				id:  supplier.ID,
			},
			wantErr: false,
		},
		{
			name: "sad :( fail to find supplier by ID - wrong supplier ID",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supplier, err := s.Supplier.FindSupplierByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SupplierUseCasesImpl.FindSupplierByID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if supplier != nil {
				if supplier.ID == "" {
					t.Errorf("expected a supplier.")
					return
				}
			}
		})
	}
}

func TestSuspendSupplier(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service")
		return
	}
	suspensionReason := `
	"This email is to inform you that as a result of your actions on April 12th, 2021, you have been issued a suspension for 1 week (7 days)"
	`
	err = s.Onboarding.UpdatePrimaryEmailAddress(ctx, primaryEmail)
	assert.Nil(t, err)

	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}

	// update biodata
	err = s.Onboarding.UpdateBioData(ctx, completeUserDetails)
	assert.Nil(t, err)

	name := "Makmende And Sons"
	partnerPractitioner := profileutils.PartnerTypePractitioner

	// Add PartnerType
	resp2, err := s.Supplier.AddPartnerType(ctx, &name, &partnerPractitioner)
	assert.Nil(t, err)
	assert.NotNil(t, resp2)
	assert.Equal(t, true, resp2)
	type args struct {
		ctx              context.Context
		suspensionReason *string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "sad case: suspend a nonexisting supplier",
			args: args{
				ctx:              context.Background(),
				suspensionReason: &suspensionReason,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Happy case: suspend an existing supplier",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Supplier.SuspendSupplier(tt.args.ctx, tt.args.suspensionReason)
			if (err != nil) != tt.wantErr {
				t.Errorf("SupplierUseCasesImpl.SuspendSupplier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SupplierUseCasesImpl.SuspendSupplier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupplierUseCasesImpl_AddPartnerType(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	testRiderName := "Test Rider"
	rider := profileutils.PartnerTypeRider
	testPractitionerName := "Test Practitioner"
	practitioner := profileutils.PartnerTypePractitioner
	testProviderName := "Test Provider"
	provider := profileutils.PartnerTypeProvider
	testPharmaceuticalName := "Test Pharmaceutical"
	pharmaceutical := profileutils.PartnerTypePharmaceutical
	testCoachName := "Test Coach"
	coach := profileutils.PartnerTypeCoach
	testNutritionName := "Test Nutrition"
	nutrition := profileutils.PartnerTypeNutrition
	testConsumerName := "Test Consumer"
	consumer := profileutils.PartnerTypeConsumer

	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service")
		return
	}
	type args struct {
		ctx         context.Context
		name        *string
		partnerType *profileutils.PartnerType
	}
	tests := []struct {
		name        string
		args        args
		want        bool
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid: add PartnerTypeRider ",
			args: args{
				ctx:         ctx,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "valid: add PartnerTypePractitioner ",
			args: args{
				ctx:         ctx,
				name:        &testPractitionerName,
				partnerType: &practitioner,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "valid: add PartnerTypeProvider ",
			args: args{
				ctx:         ctx,
				name:        &testProviderName,
				partnerType: &provider,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "valid: add PartnerTypePharmaceutical",
			args: args{
				ctx:         ctx,
				name:        &testPharmaceuticalName,
				partnerType: &pharmaceutical,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "valid: add PartnerTypeCoach",
			args: args{
				ctx:         ctx,
				name:        &testCoachName,
				partnerType: &coach,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "valid: add PartnerTypeNutrition",
			args: args{
				ctx:         ctx,
				name:        &testNutritionName,
				partnerType: &nutrition,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "invalid: add PartnerTypeConsumer",
			args: args{
				ctx:         ctx,
				name:        &testConsumerName,
				partnerType: &consumer,
			},
			want:        false,
			wantErr:     true,
			expectedErr: "invalid `partnerType`. cannot use CONSUMER in this context",
		},

		{
			name: "invalid : invalid context",
			args: args{
				ctx:         context.Background(),
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:        false,
			wantErr:     true,
			expectedErr: `unable to get the logged in user: auth token not found in context: unable to get auth token from context with key "UID" `,
		},
		{
			name: "invalid : missing name arg",
			args: args{
				ctx: ctx,
			},
			want:        false,
			wantErr:     true,
			expectedErr: "expected `name` to be defined and `partnerType` to be valid",
		},
		{
			name: "invalid: missing partnerType",
			args: args{
				ctx:  ctx,
				name: &testPharmaceuticalName,
			},
			want:        false,
			wantErr:     true,
			expectedErr: "expected `partnerType` should be valid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supplier := s
			got, err := supplier.Supplier.AddPartnerType(tt.args.ctx, tt.args.name, tt.args.partnerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SupplierUseCasesImpl.AddPartnerType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SupplierUseCasesImpl.AddPartnerType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetUpSupplier(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	individualPartner := profileutils.AccountTypeIndividual
	organizationPartner := profileutils.AccountTypeOrganisation

	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service")
		return
	}

	type args struct {
		ctx         context.Context
		accountType profileutils.AccountType
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful individual supplier account setup",
			args: args{
				ctx:         ctx,
				accountType: individualPartner,
			},
			wantErr: false,
		},
		{
			name: "Successful organization supplier account setup",
			args: args{
				ctx:         ctx,
				accountType: organizationPartner,
			},
			wantErr: false,
		},
		{
			name: "SadCase - Invalid supplier setup",
			args: args{
				ctx:         ctx,
				accountType: "non existent type",
			},
			wantErr: true,
		},
		{
			name: "SadCase - unauthenticated context",
			args: args{
				ctx:         context.Background(),
				accountType: organizationPartner,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supplier, err := s.Supplier.SetUpSupplier(tt.args.ctx, tt.args.accountType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetUpSupplier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if supplier == nil && !tt.wantErr {
				t.Errorf("expected a supplier and nil error but got: %v", err)
				return
			}

			if supplier != nil && tt.wantErr {
				t.Errorf("expected an error but instead got a nil")
				return
			}
		})
	}

}
