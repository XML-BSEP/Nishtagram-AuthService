package seeder

import (
	"auth-service/domain"
	"gorm.io/gorm"
)

func SeedData(gorm *gorm.DB) {
	gorm.Migrator().DropTable(&domain.Role{})
	gorm.Migrator().DropTable(&domain.ProfileInfo{})
	gorm.Migrator().DropTable(&domain.TotpSecret{})

	gorm.AutoMigrate(&domain.Role{})
	gorm.AutoMigrate(&domain.ProfileInfo{})
	gorm.AutoMigrate(&domain.TotpSecret{})

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
		Email: "alexignjat1998@gmail.com",
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
		ID : "43420055-3174-4c2a-9823-a8f060d644c3",
		Email: "user4@gmail.com",
		Username: "user4",
		Password: "$2y$10$OYT/DOvOVd4ofL2uWvPlbuTGU65SdyhW4vei9dqm5NxIEvrQHCf4C", //user3
		Role: roleUser,
	}
	profile5 := domain.ProfileInfo{
		ID : "ead67925-e71c-43f4-8739-c3b823fe21bb",
		Email: "user5@gmail.com",
		Username: "user5",
		Password: "$2y$10$OYT/DOvOVd4ofL2uWvPlbuTGU65SdyhW4vei9dqm5NxIEvrQHCf4C", //user3
		Role: roleUser,
	}
	profile6 := domain.ProfileInfo{
		ID : "23ddb1dd-4303-428b-b506-ff313071d5d7",
		Email: "user6@gmail.com",
		Username: "user6",
		Password: "$2y$10$OYT/DOvOVd4ofL2uWvPlbuTGU65SdyhW4vei9dqm5NxIEvrQHCf4C", //user3
		Role: roleUser,
	}
	admin := domain.ProfileInfo{
		Email: "admin1@gmail.com",
		ID : "bdb7d7c5-2c9a-4b4c-ab64-4e4828d93926",
		Username: "admin1",
		Password: "$2y$10$6KqgPNO9RrBRKCx8ZKyzKu/oorCnraEEovjMIa9FHlxRhb5tNhQOe", //admin1
		Role: roleAdmin,
	}

	agent := domain.ProfileInfo{
		Email: "agent1@gmail.com",
		ID : "1d09bb0a-d9fc-11eb-b8bc-0242ac130003",
		Username: "agent1",
		Password: "$2y$12$fbhWKmsyK8UKF28N6AKtEeyK12ziEcMI69pWSTCXcunl5fM/x31GK", //agent1
		Role: roleAgent,
	}

	gorm.Create(&profile1)
	gorm.Create(&profile2)
	gorm.Create(&profile3)
	gorm.Create(&profile4)
	gorm.Create(&profile5)
	gorm.Create(&profile6)
	gorm.Create(&admin)
	gorm.Create(&agent)

}
