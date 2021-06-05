package seeder

import (
	"auth-service/src/domain"
	"gorm.io/gorm"
)

func SeedData(gorm *gorm.DB) {
	gorm.Migrator().DropTable(&domain.Role{})
	gorm.Migrator().DropTable(&domain.ProfileInfo{})

	gorm.AutoMigrate(&domain.Role{})
	gorm.AutoMigrate(&domain.ProfileInfo{})

	seedRoles(gorm)
	seedProfiles(gorm)

}

func seedRoles(gorm *gorm.DB){
	admin := domain.Role{RoleName: "admin"}
	agent := domain.Role{RoleName: "agent"}
	user := domain.Role{RoleName: "user"}

	gorm.Create(&admin)
	gorm.Create(&agent)
	gorm.Create(&user)
}

func seedProfiles(gorm *gorm.DB) {
	var roleUser domain.Role
	var roleAdmin domain.Role
	var roleAgent domain.Role
	gorm.Where("role_name=?", "user").First(&roleUser)
	gorm.Where("role_name=?", "admin").First(&roleAdmin)
	gorm.Where("role_name=?", "agent").First(&roleAgent)

	profile1 := domain.ProfileInfo{
		ID: "e2b5f92e-c31b-11eb-8529-0242ac130003",
		Email: "user1@gmail.com",
		Username: "user1",
		Password: "$2y$10$jwbLvrZYHgZN3HFJIV1vFu.lxi6SiiKFzx2B3RItMxruVD8wNPqdS", //user1
		Role: roleUser,
	}

	profile2 := domain.ProfileInfo{
		ID : "424935b1-766c-4f99-b306-9263731518bc",
		Email: "user2@gmail.com",
		Username: "user2",
		Password: "$2y$10$D0LiWoNj3Ej7bnhq4qwX9OfQwI/zW8dJ86M0vMO0uWXw2zpmIs/r.", //user2
		Role: roleUser,
	}

	profile3 := domain.ProfileInfo{
		ID : "a2c2f993-dc32-4a82-82ed-a5f6866f7d03",
		Email: "user3@gmail.com",
		Username: "user3",
		Password: "$2y$10$OYT/DOvOVd4ofL2uWvPlbuTGU65SdyhW4vei9dqm5NxIEvrQHCf4C", //user3
		Role: roleUser,
	}

	profile4 := domain.ProfileInfo{
		Email: "admin1@gmail.com",
		ID : "43420055-3174-4c2a-9823-a8f060d644c3",
		Username: "user4",
		Password: "$2y$10$6KqgPNO9RrBRKCx8ZKyzKu/oorCnraEEovjMIa9FHlxRhb5tNhQOe", //admin1
		Role: roleUser,
	}

	gorm.Create(&profile1)
	gorm.Create(&profile2)
	gorm.Create(&profile3)
	gorm.Create(&profile4)

}
