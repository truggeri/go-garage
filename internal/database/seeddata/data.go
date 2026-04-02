package seeddata

import (
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// GetSampleUsers returns a list of sample users for seeding the database
// Password hashes are bcrypt hashes for "password123"
// Note: In production, each user should have a unique password
func GetSampleUsers() []*models.User {
	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)

	return []*models.User{
		{
			ID:           "550e8400-e29b-41d4-a716-446655440001",
			Username:     "john_doe",
			Email:        "john.doe@example.com",
			PasswordHash: "$2a$10$Q0cA5lEcp0O9jhRKD5N2b.hWnSN5oujuSlPnAKag60TaiK9avvlB.", // password123
			FirstName:    "John",
			LastName:     "Doe",
			CreatedAt:    now,
			UpdatedAt:    now,
			LastLoginAt:  &lastLogin,
		},
		{
			ID:           "550e8400-e29b-41d4-a716-446655440002",
			Username:     "jane_smith",
			Email:        "jane.smith@example.com",
			PasswordHash: "$2a$10$Q0cA5lEcp0O9jhRKD5N2b.hWnSN5oujuSlPnAKag60TaiK9avvlB.", // password123
			FirstName:    "Jane",
			LastName:     "Smith",
			CreatedAt:    now,
			UpdatedAt:    now,
			LastLoginAt:  nil,
		},
		{
			ID:           "550e8400-e29b-41d4-a716-446655440003",
			Username:     "bob_wilson",
			Email:        "bob.wilson@example.com",
			PasswordHash: "$2a$10$Q0cA5lEcp0O9jhRKD5N2b.hWnSN5oujuSlPnAKag60TaiK9avvlB.", // password123
			FirstName:    "Robert",
			LastName:     "Wilson",
			CreatedAt:    now,
			UpdatedAt:    now,
			LastLoginAt:  nil,
		},
	}
}

// GetSampleVehicles returns a list of sample vehicles for seeding the database
// These vehicles are linked to the sample users
func GetSampleVehicles() []*models.Vehicle {
	now := time.Now()
	purchaseDate1 := now.Add(-365 * 24 * time.Hour) // 1 year ago
	purchaseDate2 := now.Add(-730 * 24 * time.Hour) // 2 years ago
	purchaseDate3 := now.Add(-180 * 24 * time.Hour) // 6 months ago
	purchaseDate4 := now.Add(-90 * 24 * time.Hour)  // 3 months ago

	price1 := 25000.00
	price2 := 35000.00
	price3 := 18000.00
	price4 := 42000.00
	price5 := 28000.00

	mileage1 := 50000
	mileage2 := 75000
	mileage3 := 30000
	mileage4 := 15000
	mileage5 := 45000

	currentMileage1 := 65000
	currentMileage2 := 85000
	currentMileage3 := 35000
	currentMileage4 := 20000
	currentMileage5 := 50000

	return []*models.Vehicle{
		{
			ID:              "660e8400-e29b-41d4-a716-446655440001",
			UserID:          "550e8400-e29b-41d4-a716-446655440001", // john_doe
			VIN:             "1HGBH41JXMN109186",
			Make:            "Honda",
			Model:           "Accord",
			Year:            2020,
			Color:           "Silver",
			LicensePlate:    "ABC123",
			PurchaseDate:    &purchaseDate1,
			PurchasePrice:   &price1,
			PurchaseMileage: &mileage1,
			CurrentMileage:  &currentMileage1,
			Status:          models.VehicleStatusActive,
			Notes:           "Daily driver, excellent condition",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440002",
			UserID:          "550e8400-e29b-41d4-a716-446655440001", // john_doe
			VIN:             "2FMDK3GC5BBB02514",
			Make:            "Ford",
			Model:           "Edge",
			Year:            2019,
			Color:           "Blue",
			LicensePlate:    "XYZ789",
			PurchaseDate:    &purchaseDate2,
			PurchasePrice:   &price2,
			PurchaseMileage: &mileage2,
			CurrentMileage:  &currentMileage2,
			Status:          models.VehicleStatusActive,
			Notes:           "Family SUV",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440003",
			UserID:          "550e8400-e29b-41d4-a716-446655440002", // jane_smith
			VIN:             "5YFBURHE5HP584912",
			Make:            "Toyota",
			Model:           "Corolla",
			Year:            2021,
			Color:           "White",
			LicensePlate:    "DEF456",
			PurchaseDate:    &purchaseDate3,
			PurchasePrice:   &price3,
			PurchaseMileage: &mileage3,
			CurrentMileage:  &currentMileage3,
			Status:          models.VehicleStatusActive,
			Notes:           "Fuel efficient commuter car",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440004",
			UserID:          "550e8400-e29b-41d4-a716-446655440002", // jane_smith
			VIN:             "3VW2B7AJ9DM456789",
			Make:            "Volkswagen",
			Model:           "Jetta",
			Year:            2018,
			Color:           "Black",
			LicensePlate:    "GHI789",
			PurchaseDate:    nil,
			PurchasePrice:   nil,
			PurchaseMileage: nil,
			CurrentMileage:  nil,
			Status:          models.VehicleStatusSold,
			Notes:           "Sold in 2023",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440005",
			UserID:          "550e8400-e29b-41d4-a716-446655440003", // bob_wilson
			VIN:             "1G1ZD5ST8LF123456",
			Make:            "Chevrolet",
			Model:           "Malibu",
			Year:            2022,
			Color:           "Red",
			LicensePlate:    "JKL012",
			PurchaseDate:    &purchaseDate4,
			PurchasePrice:   &price4,
			PurchaseMileage: &mileage4,
			CurrentMileage:  &currentMileage4,
			Status:          models.VehicleStatusActive,
			Notes:           "New car with warranty",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440006",
			UserID:          "550e8400-e29b-41d4-a716-446655440003", // bob_wilson
			VIN:             "WBAPH7C58BE123456",
			Make:            "BMW",
			Model:           "3 Series",
			Year:            2015,
			Color:           "Gray",
			LicensePlate:    "MNO345",
			PurchaseDate:    &purchaseDate2,
			PurchasePrice:   &price5,
			PurchaseMileage: &mileage5,
			CurrentMileage:  &currentMileage5,
			Status:          models.VehicleStatusScrapped,
			Notes:           "Totaled in accident, scrapped",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
}

// GetSampleMaintenanceRecords returns a list of sample maintenance records for seeding the database
// These records are linked to the sample vehicles
func GetSampleMaintenanceRecords() []*models.MaintenanceRecord {
	now := time.Now()
	serviceDate1 := now.Add(-30 * 24 * time.Hour)  // 1 month ago
	serviceDate2 := now.Add(-60 * 24 * time.Hour)  // 2 months ago
	serviceDate3 := now.Add(-90 * 24 * time.Hour)  // 3 months ago
	serviceDate4 := now.Add(-120 * 24 * time.Hour) // 4 months ago
	serviceDate5 := now.Add(-150 * 24 * time.Hour) // 5 months ago
	serviceDate6 := now.Add(-180 * 24 * time.Hour) // 6 months ago
	serviceDate7 := now.Add(-210 * 24 * time.Hour) // 7 months ago
	serviceDate8 := now.Add(-240 * 24 * time.Hour) // 8 months ago

	cost1 := 45.99
	cost2 := 89.50
	cost3 := 350.00
	cost4 := 125.00
	cost5 := 55.99
	cost6 := 450.00
	cost7 := 75.00
	cost8 := 220.00

	mileage1 := 64500
	mileage2 := 63000
	mileage3 := 61000
	mileage4 := 84000
	mileage5 := 83000
	mileage6 := 34000
	mileage7 := 33000
	mileage8 := 19000

	return []*models.MaintenanceRecord{
		{
			ID:               "770e8400-e29b-41d4-a716-446655440001",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			ServiceType:      "oil_change",
			ServiceDate:      serviceDate1,
			MileageAtService: &mileage1,
			Cost:             &cost1,
			ServiceProvider:  "Quick Lube",
			Notes:            "Synthetic oil, replaced air filter",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440002",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			ServiceType:      "tire_rotation",
			ServiceDate:      serviceDate2,
			MileageAtService: &mileage2,
			Cost:             &cost2,
			ServiceProvider:  "Tire World",
			Notes:            "Rotated and balanced all four tires",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440003",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			ServiceType:      "brakes",
			ServiceDate:      serviceDate3,
			MileageAtService: &mileage3,
			Cost:             &cost3,
			ServiceProvider:  "Brake Masters",
			Notes:            "Replaced front brake pads and rotors",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440004",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440002", // Ford Edge
			ServiceType:      "oil_change",
			ServiceDate:      serviceDate4,
			MileageAtService: &mileage4,
			Cost:             &cost4,
			ServiceProvider:  "Ford Dealership",
			Notes:            "Full synthetic oil change with inspection",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440005",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440002", // Ford Edge
			ServiceType:      "glass",
			ServiceDate:      serviceDate5,
			MileageAtService: &mileage5,
			Cost:             &cost5,
			ServiceProvider:  "State Inspection Center",
			Notes:            "Annual safety inspection - passed",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440006",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440003", // Toyota Corolla
			ServiceType:      "brakes",
			ServiceDate:      serviceDate6,
			MileageAtService: &mileage6,
			Cost:             &cost6,
			ServiceProvider:  "Toyota Service Center",
			Notes:            "Replaced all brake pads and rear rotors",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440007",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440003", // Toyota Corolla
			ServiceType:      "tire_rotation",
			ServiceDate:      serviceDate7,
			MileageAtService: &mileage7,
			Cost:             &cost7,
			ServiceProvider:  "Toyota Service Center",
			Notes:            "Tire rotation and pressure check",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "770e8400-e29b-41d4-a716-446655440008",
			VehicleID:        "660e8400-e29b-41d4-a716-446655440005", // Chevrolet Malibu
			ServiceType:      "oil_change",
			ServiceDate:      serviceDate8,
			MileageAtService: &mileage8,
			Cost:             &cost8,
			ServiceProvider:  "Chevrolet Dealership",
			Notes:            "First oil change, complimentary service",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
}

// GetSampleFuelRecords returns a list of sample fuel records for seeding the database
// These records are linked to the sample vehicles
func GetSampleFuelRecords() []*models.FuelRecord {
	now := time.Now()
	fillDate1 := now.Add(-7 * 24 * time.Hour)  // 1 week ago
	fillDate2 := now.Add(-21 * 24 * time.Hour) // 3 weeks ago
	fillDate3 := now.Add(-35 * 24 * time.Hour) // 5 weeks ago
	fillDate4 := now.Add(-14 * 24 * time.Hour) // 2 weeks ago
	fillDate5 := now.Add(-28 * 24 * time.Hour) // 4 weeks ago
	fillDate6 := now.Add(-10 * 24 * time.Hour) // 10 days ago
	fillDate7 := now.Add(-42 * 24 * time.Hour) // 6 weeks ago
	fillDate8 := now.Add(-56 * 24 * time.Hour) // 8 weeks ago

	price1 := 3.459
	price2 := 3.299
	price3 := 3.529
	price4 := 3.899
	price5 := 3.199
	price6 := 3.749
	price7 := 4.099
	price8 := 3.359

	octane1 := 87
	octane2 := 87
	octane3 := 91
	octane6 := 93

	cityPct1 := 70
	cityPct3 := 40
	cityPct5 := 80
	cityPct7 := 50

	mpg1 := 28.5
	mpg4 := 24.2
	mpg6 := 30.1

	return []*models.FuelRecord{
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440001",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			FillDate:              fillDate1,
			Mileage:               64800,
			Volume:                12.5,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price1,
			OctaneRating:          &octane1,
			Location:              "Shell Station, Main St",
			Brand:                 "Shell",
			Notes:                 "Regular fill-up",
			CityDrivingPercentage: &cityPct1,
			VehicleReportedMPG:    &mpg1,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440002",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			FillDate:              fillDate2,
			Mileage:               64450,
			Volume:                11.8,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price2,
			OctaneRating:          &octane2,
			Location:              "Costco Gas, Oak Ave",
			Brand:                 "Kirkland",
			Notes:                 "",
			CityDrivingPercentage: nil,
			VehicleReportedMPG:    nil,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440003",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440001", // Honda Accord
			FillDate:              fillDate3,
			Mileage:               64100,
			Volume:                13.2,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price3,
			OctaneRating:          &octane3,
			Location:              "BP Station, Elm Blvd",
			Brand:                 "BP",
			Notes:                 "Premium fuel for highway trip",
			CityDrivingPercentage: &cityPct3,
			VehicleReportedMPG:    nil,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440004",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440002", // Ford Edge
			FillDate:              fillDate4,
			Mileage:               84500,
			Volume:                16.3,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price4,
			OctaneRating:          nil,
			Location:              "Exxon, Highway 101",
			Brand:                 "Exxon",
			Notes:                 "Long road trip fill-up",
			CityDrivingPercentage: nil,
			VehicleReportedMPG:    &mpg4,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440005",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440003", // Toyota Corolla
			FillDate:              fillDate5,
			Mileage:               34500,
			Volume:                10.1,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price5,
			OctaneRating:          nil,
			Location:              "Chevron, Pine St",
			Brand:                 "Chevron",
			Notes:                 "Great mileage this tank",
			CityDrivingPercentage: &cityPct5,
			VehicleReportedMPG:    nil,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440006",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440003", // Toyota Corolla
			FillDate:              fillDate6,
			Mileage:               34800,
			Volume:                9.8,
			FuelType:              "gasoline",
			PartialFill:           true,
			PricePerUnit:          &price6,
			OctaneRating:          &octane6,
			Location:              "Mobil, Cedar Rd",
			Brand:                 "Mobil",
			Notes:                 "Partial fill, was in a hurry",
			CityDrivingPercentage: nil,
			VehicleReportedMPG:    &mpg6,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440007",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440005", // Chevrolet Malibu
			FillDate:              fillDate7,
			Mileage:               19500,
			Volume:                13.7,
			FuelType:              "gasoline",
			PartialFill:           false,
			PricePerUnit:          &price7,
			OctaneRating:          nil,
			Location:              "Sunoco, Market St",
			Brand:                 "Sunoco",
			Notes:                 "",
			CityDrivingPercentage: &cityPct7,
			VehicleReportedMPG:    nil,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "880e8400-e29b-41d4-a716-446655440008",
			VehicleID:             "660e8400-e29b-41d4-a716-446655440005", // Chevrolet Malibu
			FillDate:              fillDate8,
			Mileage:               19100,
			Volume:                14.0,
			FuelType:              "e85",
			PartialFill:           false,
			PricePerUnit:          &price8,
			OctaneRating:          nil,
			Location:              "FlexFuel Station, 5th Ave",
			Brand:                 "FlexFuel",
			Notes:                 "Trying E85 fuel",
			CityDrivingPercentage: nil,
			VehicleReportedMPG:    nil,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
	}
}
