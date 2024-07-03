package test

import (
	"AuthService/internal/pb"
	"AuthService/test/testsuite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, ts := testsuite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := ts.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.UserId)

	respLogin, err := ts.AuthClient.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.Cfg.JWTSecretKey), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.UserId, int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(24*180*time.Hour).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := testsuite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.UserId)

	respReg, err = st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.UserId)
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := testsuite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := testsuite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			expectedErr: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &pb.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidate_FailCases(t *testing.T) {
	ctx, ts := testsuite.New(t)

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = gofakeit.Uint64()
	claims["email"] = gofakeit.Email()
	claims["exp"] = time.Now().Local().Add(-1 * time.Hour).Unix()
	claims["issuer"] = "go-grpc-auth-svc"
	tokenString, _ := token.SignedString([]byte(ts.Cfg.JWTSecretKey))

	tests := []struct {
		name        string
		token       string
		expectedErr string
	}{
		{
			name:        "Login with Empty token",
			token:       "",
			expectedErr: "token is required",
		},
		{
			name:        "Login with Wrong token",
			token:       "example wrong jwt",
			expectedErr: "invalid JWT",
		},
		{
			name:        "Login with Expired token",
			token:       tokenString,
			expectedErr: "invalid JWT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ts.AuthClient.Validate(ctx, &pb.ValidateRequest{
				Token: tt.token,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, 10)
}
