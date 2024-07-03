package utils

import (
	"AuthService/internal/models"
	"AuthService/internal/pb"
)

func ConvertUsers(users []*models.UserDTO) []*pb.Student {
	var pbStudents []*pb.Student
	for _, user := range users {
		pbStudents = append(pbStudents, &pb.Student{
			Name:     user.Name,
			Lastname: user.Lastname,
		})
	}
	return pbStudents
}
