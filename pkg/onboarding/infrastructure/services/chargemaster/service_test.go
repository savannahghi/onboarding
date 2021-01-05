package chargemaster

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.slade360emr.com/go/base"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/resources"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/domain"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/infrastructure/database"
)

func TestServiceChargeMasterImpl_FindProvider(t *testing.T) {
	fr, err := database.NewFirebaseRepository(context.Background())
	if err != nil {
		t.Errorf("can't instantiate firebase repository in resolver: %w", err)
		return
	}
	cm := NewChargeMasterUseCasesImpl(fr)
	assert.NotNil(t, cm)
	type args struct {
		ctx        context.Context
		pagination *base.PaginationInput
		filter     []*resources.BusinessPartnerFilterInput
		sort       []*resources.BusinessPartnerSortInput
	}
	first := 10
	after := "0"
	last := 10
	before := "20"
	testSladeCode := "PRO-50"
	ascSort := base.SortOrderAsc
	invalidPage := "invalidpage"

	tests := []struct {
		name                   string
		args                   args
		wantErr                bool
		expectNonNilConnection bool
		expectedErr            error
	}{
		{
			name:                   "happy case - query params only no pagination filter or sort params",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter:     []*resources.BusinessPartnerFilterInput{},
				sort:       []*resources.BusinessPartnerSortInput{},
			},
		},
		{
			name:                   "happy case - with forward pagination",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					First: first,
					After: after,
				},
				filter: []*resources.BusinessPartnerFilterInput{},
				sort:   []*resources.BusinessPartnerSortInput{},
			},
		},
		{
			name:                   "happy case - with backward pagination",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					Last:   last,
					Before: before,
				},
				filter: []*resources.BusinessPartnerFilterInput{},
				sort:   []*resources.BusinessPartnerSortInput{},
			},
		},
		{
			name:                   "happy case - with filter",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter: []*resources.BusinessPartnerFilterInput{
					{
						SladeCode: &testSladeCode,
					},
				},
				sort: []*resources.BusinessPartnerSortInput{},
			},
		},
		{
			name:                   "happy case - with sort",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter:     []*resources.BusinessPartnerFilterInput{},
				sort: []*resources.BusinessPartnerSortInput{
					{
						Name:      &ascSort,
						SladeCode: &ascSort,
					},
				},
			},
		},
		{
			name:                   "sad case - with invalid pagination",
			expectNonNilConnection: false,
			expectedErr:            errors.New("expected `after` to be parseable as an int; got invalidpage"),
			wantErr:                true,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					After: invalidPage,
				},
				filter: []*resources.BusinessPartnerFilterInput{},
				sort:   []*resources.BusinessPartnerSortInput{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cm.FindProvider(tt.args.ctx, tt.args.pagination, tt.args.filter, tt.args.sort)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceChargeMasterImpl.FindProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expectNonNilConnection {
				assert.NotNil(t, got)
			}
			if tt.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			}
		})
	}
}

func TestServiceChargeMasterImpl_FindBranch(t *testing.T) {
	fr, err := database.NewFirebaseRepository(context.Background())
	if err != nil {
		t.Errorf("can't instantiate firebase repository in resolver: %w", err)
		return
	}
	cm := NewChargeMasterUseCasesImpl(fr)
	assert.NotNil(t, cm)
	type args struct {
		ctx        context.Context
		pagination *base.PaginationInput
		filter     []*resources.BranchFilterInput
		sort       []*resources.BranchSortInput
	}
	first := 10
	after := "0"
	last := 10
	before := "20"
	testSladeCode := "PRO-50"
	ascSort := base.SortOrderAsc
	invalidPage := "invalidpage"

	tests := []struct {
		name                   string
		args                   args
		wantErr                bool
		expectNonNilConnection bool
		expectedErr            error
	}{
		{
			name:                   "happy case - query params only no pagination filter or sort params",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter:     []*resources.BranchFilterInput{},
				sort:       []*resources.BranchSortInput{},
			},
		},
		{
			name:                   "happy case - with forward pagination",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					First: first,
					After: after,
				},
				filter: []*resources.BranchFilterInput{},
				sort:   []*resources.BranchSortInput{},
			},
		},
		{
			name:                   "happy case - with backward pagination",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					Last:   last,
					Before: before,
				},
				filter: []*resources.BranchFilterInput{},
				sort:   []*resources.BranchSortInput{},
			},
		},
		{
			name:                   "happy case -with filter",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter: []*resources.BranchFilterInput{
					{
						SladeCode: &testSladeCode,
					},
				},
				sort: []*resources.BranchSortInput{},
			},
		},
		{
			name:                   "happy case - with sort",
			expectNonNilConnection: true,
			expectedErr:            nil,
			wantErr:                false,
			args: args{
				ctx:        context.Background(),
				pagination: &base.PaginationInput{},
				filter:     []*resources.BranchFilterInput{},
				sort: []*resources.BranchSortInput{
					{
						Name:      &ascSort,
						SladeCode: &ascSort,
					},
				},
			},
		},
		{
			name:                   "sad case - with invalid pagination",
			expectNonNilConnection: false,
			expectedErr:            errors.New("expected `after` to be parseable as an int; got invalidpage"),
			wantErr:                true,
			args: args{
				ctx: context.Background(),
				pagination: &base.PaginationInput{
					After: invalidPage,
				},
				filter: []*resources.BranchFilterInput{},
				sort:   []*resources.BranchSortInput{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cm.FindBranch(tt.args.ctx, tt.args.pagination, tt.args.filter, tt.args.sort)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceChargeMasterImpl.FindBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expectNonNilConnection {
				assert.NotNil(t, got)
			}
			if tt.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			}
		})
	}
}

func Test_parentOrgSladeCodeFromBranch(t *testing.T) {
	type args struct {
		branch *domain.BusinessPartner
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				branch: &domain.BusinessPartner{
					SladeCode: "BRA-PRO-4313-1",
				},
			},
			want:    "PRO-4313",
			wantErr: false,
		},
		{
			name: "no BRA prefix",
			args: args{
				branch: &domain.BusinessPartner{
					SladeCode: "PRO-4313-1",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "sad case (long branch slade code)",
			args: args{
				branch: &domain.BusinessPartner{
					SladeCode: "BRA-PRO-4313-1-9393030",
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parentOrgSladeCodeFromBranch(tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("parentOrgSladeCodeFromBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parentOrgSladeCodeFromBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}