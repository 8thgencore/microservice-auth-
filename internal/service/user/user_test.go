package user

import (
	"context"
	"database/sql"
	"testing"

	"github.com/8thgencore/microservice-auth/internal/model"
	"github.com/8thgencore/microservice-auth/internal/repository"
	"github.com/8thgencore/microservice-common/pkg/db"
	"github.com/8thgencore/microservice-common/pkg/db/transaction"
	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	repositoryMocks "github.com/8thgencore/microservice-auth/internal/repository/mocks"
	dbMocks "github.com/8thgencore/microservice-common/pkg/db/mocks"
)

var (
	id              = "uuid"
	name            = "name"
	email           = "email"
	password        = "password"
	passwordConfirm = "passwordConfirm"
	role            = "USER"
	createdAt       = timestamppb.Now()
	updatedAt       = timestamppb.Now()

	user = &model.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: createdAt.AsTime(),
		UpdatedAt: sql.NullTime{
			Time:  updatedAt.AsTime(),
			Valid: true,
		},
	}
)

var (
	opts = pgx.TxOptions{IsoLevel: pgx.ReadCommitted}

	transactorCommitMock = func(mc *minimock.Controller) db.Transactor {
		mock := dbMocks.NewTransactorMock(mc)
		txMock := dbMocks.NewTxMock(mc)
		mock.BeginTxMock.Expect(minimock.AnyContext, opts).Return(txMock, nil)
		txMock.CommitMock.Expect(minimock.AnyContext).Return(nil)
		return mock
	}

	transactorRollbackMock = func(mc *minimock.Controller) db.Transactor {
		mock := dbMocks.NewTransactorMock(mc)
		txMock := dbMocks.NewTxMock(mc)
		mock.BeginTxMock.Expect(minimock.AnyContext, opts).Return(txMock, nil)
		txMock.RollbackMock.Expect(minimock.AnyContext).Return(nil)
		return mock
	}
)

// TestCreate tests the creation of a new user.
func TestCreate(t *testing.T) {
	t.Parallel()

	type userRepositoryMockFunc func(mc *minimock.Controller) repository.UserRepository
	type logRepositoryMockFunc func(mc *minimock.Controller) repository.LogRepository
	type transactorMockFunc func(mc *minimock.Controller) db.Transactor

	type args struct {
		ctx context.Context
		req *model.UserCreate
	}

	var (
		ctx = context.Background()

		req = &model.UserCreate{
			Name:            name,
			Email:           email,
			Password:        password,
			PasswordConfirm: password,
			Role:            role,
		}

		reqPassNotMatch = &model.UserCreate{
			Name:            name,
			Email:           email,
			Password:        password,
			PasswordConfirm: passwordConfirm,
			Role:            role,
		}
	)

	tests := []struct {
		name               string
		args               args
		want               string
		err                error
		userRepositoryMock userRepositoryMockFunc
		logRepositoryMock  logRepositoryMockFunc
		transactorMock     transactorMockFunc
	}{
		{
			name: "passwords match error case",
			args: args{
				ctx: ctx,
				req: reqPassNotMatch,
			},
			want: "",
			err:  ErrPasswordsMismatch,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: func(mc *minimock.Controller) db.Transactor {
				mock := dbMocks.NewTransactorMock(mc)
				return mock
			},
		},
		{
			name: "user repository error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: "",
			err:  ErrUserCreate,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.CreateMock.Optional().Return("", ErrUserCreate)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},

		{
			name: "log repository error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: "",
			err:  ErrUserCreate,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.CreateMock.Optional().Return(user.ID, nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(ErrUserCreate)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},

		{
			name: "user with existing name",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: "",
			err:  ErrUserNameExists,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.CreateMock.Optional().Return("", ErrUserNameExists)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},

		{
			name: "user with existing email",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: "",
			err:  ErrUserEmailExists,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.CreateMock.Optional().Return("", ErrUserEmailExists)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},

		{
			name: "success case",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: id,
			err:  nil,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.CreateMock.Optional().Return(id, nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(nil)
				return mock
			},
			transactorMock: transactorCommitMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			userRepositoryMock := tt.userRepositoryMock(mc)
			logRepositoryMock := tt.logRepositoryMock(mc)
			txManagerMock := transaction.NewTransactionManager(tt.transactorMock(mc))
			srv := NewService(userRepositoryMock, logRepositoryMock, txManagerMock)

			user := &model.UserCreate{}
			copier.Copy(&user, &tt.args.req)

			res, err := srv.Create(tt.args.ctx, user)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, res)
		})
	}
}

// TestGet tests the retrieval of an existing user.
func TestGet(t *testing.T) {
	t.Parallel()
	type userRepositoryMockFunc func(mc *minimock.Controller) repository.UserRepository
	type logRepositoryMockFunc func(mc *minimock.Controller) repository.LogRepository
	type transactorMockFunc func(mc *minimock.Controller) db.Transactor

	type args struct {
		ctx context.Context
		req string
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)
	)

	tests := []struct {
		name               string
		args               args
		want               *model.User
		err                error
		userRepositoryMock userRepositoryMockFunc
		logRepositoryMock  logRepositoryMockFunc
		transactorMock     transactorMockFunc
	}{
		{
			name: "success case",
			args: args{
				ctx: ctx,
				req: id,
			},
			want: user,
			err:  nil,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(user, nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(nil)
				return mock
			},
			transactorMock: transactorCommitMock,
		},
		{
			name: "user repository get error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			want: nil,
			err:  ErrUserRead,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(nil, ErrUserRead)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "log repository error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			want: nil,
			err:  ErrUserRead,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(user, nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(ErrUserRead)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "user not found error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			want: nil,
			err:  ErrUserNotFound,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(nil, ErrUserNotFound)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepositoryMock := tt.userRepositoryMock(mc)
			logRepositoryMock := tt.logRepositoryMock(mc)
			txManagerMock := transaction.NewTransactionManager(tt.transactorMock(mc))
			srv := NewService(userRepositoryMock, logRepositoryMock, txManagerMock)

			res, err := srv.Get(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, res)
		})
	}
}

// TestUpdate tests the update of an existing user.
func TestUpdate(t *testing.T) {
	t.Parallel()
	type userRepositoryMockFunc func(mc *minimock.Controller) repository.UserRepository
	type logRepositoryMockFunc func(mc *minimock.Controller) repository.LogRepository
	type transactorMockFunc func(mc *minimock.Controller) db.Transactor

	type args struct {
		ctx context.Context
		req *model.UserUpdate
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		req = &model.UserUpdate{
			ID: id,
			Name: sql.NullString{
				String: name,
				Valid:  true,
			},
			Email: sql.NullString{
				String: email,
				Valid:  true,
			},
			Role: sql.NullString{
				String: role,
				Valid:  true,
			},
		}
	)

	tests := []struct {
		name               string
		args               args
		err                error
		userRepositoryMock userRepositoryMockFunc
		logRepositoryMock  logRepositoryMockFunc
		transactorMock     transactorMockFunc
	}{
		{
			name: "success case",
			args: args{
				ctx: ctx,
				req: req,
			},
			err: nil,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.UpdateMock.Expect(minimock.AnyContext, req).Return(nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(nil)
				return mock
			},
			transactorMock: transactorCommitMock,
		},
		{
			name: "user repository get error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			err: ErrUserUpdate,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(nil, ErrUserUpdate)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "user repository update error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			err: ErrUserUpdate,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.UpdateMock.Expect(minimock.AnyContext, req).Return(ErrUserUpdate)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "log repository error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			err: ErrUserUpdate,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.UpdateMock.Expect(minimock.AnyContext, req).Return(nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(ErrUserUpdate)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepositoryMock := tt.userRepositoryMock(mc)
			logRepositoryMock := tt.logRepositoryMock(mc)
			txManagerMock := transaction.NewTransactionManager(tt.transactorMock(mc))
			srv := NewService(userRepositoryMock, logRepositoryMock, txManagerMock)

			err := srv.Update(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.err, err)
		})
	}
}

// TestDelete tests the deletion of an existing user.
func TestDelete(t *testing.T) {
	t.Parallel()
	type userRepositoryMockFunc func(mc *minimock.Controller) repository.UserRepository
	type logRepositoryMockFunc func(mc *minimock.Controller) repository.LogRepository
	type transactorMockFunc func(mc *minimock.Controller) db.Transactor

	type args struct {
		ctx context.Context
		req string
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)
	)

	tests := []struct {
		name               string
		args               args
		err                error
		userRepositoryMock userRepositoryMockFunc
		logRepositoryMock  logRepositoryMockFunc
		transactorMock     transactorMockFunc
	}{
		{
			name: "success case",
			args: args{
				ctx: ctx,
				req: id,
			},
			err: nil,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.DeleteMock.Expect(minimock.AnyContext, id).Return(nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(nil)
				return mock
			},
			transactorMock: transactorCommitMock,
		},
		{
			name: "user repository get error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			err: ErrUserDelete,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(nil, ErrUserDelete)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "user repository delete error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			err: ErrUserDelete,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.DeleteMock.Expect(minimock.AnyContext, id).Return(ErrUserDelete)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "log repository error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			err: ErrUserDelete,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, user.ID).Return(user, nil)
				mock.DeleteMock.Expect(minimock.AnyContext, id).Return(nil)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				mock.LogMock.Optional().Return(ErrUserDelete)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
		{
			name: "user not found error case",
			args: args{
				ctx: ctx,
				req: id,
			},
			err: ErrUserNotFound,
			userRepositoryMock: func(mc *minimock.Controller) repository.UserRepository {
				mock := repositoryMocks.NewUserRepositoryMock(mc)
				mock.GetMock.Expect(minimock.AnyContext, id).Return(nil, ErrUserNotFound)
				return mock
			},
			logRepositoryMock: func(mc *minimock.Controller) repository.LogRepository {
				mock := repositoryMocks.NewLogRepositoryMock(mc)
				return mock
			},
			transactorMock: transactorRollbackMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepositoryMock := tt.userRepositoryMock(mc)
			logRepositoryMock := tt.logRepositoryMock(mc)
			txManagerMock := transaction.NewTransactionManager(tt.transactorMock(mc))
			srv := NewService(userRepositoryMock, logRepositoryMock, txManagerMock)

			err := srv.Delete(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.err, err)
		})
	}
}
