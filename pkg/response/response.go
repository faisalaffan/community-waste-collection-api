package response

import "github.com/gofiber/fiber/v3"

type envelope struct {
	Status     string      `json:"status"`
	Data       interface{} `json:"data,omitempty"`
	Error_     *apiError   `json:"error,omitempty"`
	Pagination *pagination `json:"pagination,omitempty"`
}

type apiError struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func Success(c fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(envelope{Status: "success", Data: data})
}

func SuccessCreated(c fiber.Ctx, data interface{}) error {
	return Success(c, 201, data)
}

func SuccessOK(c fiber.Ctx, data interface{}) error {
	return Success(c, 200, data)
}

func SuccessPaginated(c fiber.Ctx, data interface{}, page, perPage int, total int64) error {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	return c.Status(200).JSON(envelope{
		Status: "success",
		Data:   data,
		Pagination: &pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func Error(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(envelope{
		Status: "error",
		Error_: &apiError{Code: code, Message: message},
	})
}

func ValidationError(c fiber.Ctx, message string, details []ValidationDetail) error {
	return c.Status(422).JSON(envelope{
		Status: "fail",
		Error_: &apiError{Code: "VALIDATION_ERROR", Message: message, Details: details},
	})
}
