package user

import (
	"context"
	"github.com/JoaoRafa19/gobid/internal/validator"
)

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (c CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator = make(validator.Evaluator)

	eval.CheckField(validator.NotBlank(c.UserName), "user_name", "user name cannot be blank")
	eval.CheckField(validator.NotBlank(c.Email), "email", "email cannot be blank")
	eval.CheckField(validator.NotBlank(c.Bio), "bio", "bio cannot be blank")
	eval.CheckField(validator.MinChar(c.Password, 8), "password", "password cannot be less than 8 characters")
	eval.CheckField(validator.Matches(c.Email, validator.EmailRX), "email", "email is not valid")
	eval.CheckField(
		validator.MinChar(c.Bio, 10) && validator.MaxChar(c.Bio, 255),
		"bio",
		"bio cannot be less than 10 characters",
	)
	return eval

}
