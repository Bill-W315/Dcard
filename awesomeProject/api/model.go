package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"strings"
	"time"
)

var redisClient *redis.Client
var dbClient *sql.DB

func setConnections() {
	redisClient = connectRedis()
	dbClient = connectDatabase()
}

func connectRedis() *redis.Client {
	options := redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 1000,
	}
	return redis.NewClient(&options)
}

func connectDatabase() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:bill880315@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(90)
	db.SetMaxOpenConns(90)
	return db
}

func clearSearchHistory() {
	err := redisClient.FlushAll(context.Background()).Err()
	if err != nil {
		println(err.Error())
		return
	}
}

func saveAd(ad Ad) error {
	//check empty list
	if len(ad.Conditions.Gender) == 0 {
		ad.Conditions.Gender = nil
	}
	if len(ad.Conditions.Countries) == 0 {
		ad.Conditions.Countries = nil
	}
	if len(ad.Conditions.Platforms) == 0 {
		ad.Conditions.Platforms = nil
	}

	newUUID := uuid.New()
	countryJson, err := json.Marshal(ad.Conditions.Countries)
	platformsJson, err := json.Marshal(ad.Conditions.Platforms)
	genderJson, err := json.Marshal(ad.Conditions.Gender)
	query := "INSERT INTO ad (uuid, title, start_at, end_at, age_start, age_end, Country, Platform, Gender) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	_, err = dbClient.Query(query, newUUID, ad.Title, ad.StartAt, ad.EndAt, ad.Conditions.AgeStart, ad.Conditions.AgeEnd, string(countryJson), string(platformsJson), string(genderJson))
	if err != nil {
		fmt.Println("Error save ad to database: ", err)
		return err
	}

	return nil
}

func getAdsByConditions(condition SearchCondition) ([]SearchResult, error) {

	var resultAds = []SearchResult{}
	var tmpAds = []Ad{}
	ctx := context.Background()

	//First check if param combination is in cache
	conditionStr, err := json.Marshal(condition)
	if err != nil {
		return nil, errors.New("Cannot parse condition into JSON string!")
	}
	cacheResult := redisClient.Get(ctx, string(conditionStr)).Val()
	//If param combination exists,cache will return ads
	if cacheResult != "" {
		err := json.Unmarshal([]byte(cacheResult), &tmpAds)
		if err != nil {
			return nil, err
		}
	} else {
		//If not, search ad by condition and add to cache
		tmpAds = getAdsByCondition(condition)
		adsJson, err := json.Marshal(tmpAds)
		if err != nil {
			return nil, err
		}
		//Save to cache and set expire time by closest end time to now
		if len(tmpAds) > 0 {
			minTime := tmpAds[0].EndAt
			for _, ad := range tmpAds {
				if ad.EndAt.Before(minTime) {
					minTime = ad.EndAt
				}
			}
			err = redisClient.Set(ctx, string(conditionStr), adsJson, minTime.Sub(getNowTime())).Err()
			if err != nil {
				return nil, err
			}
		} else {
			err = redisClient.Set(ctx, string(conditionStr), adsJson, 10*time.Second).Err()
			if err != nil {
				println(err.Error())
				return nil, err
			}
		}
	}

	//Pagination
	//Offset > Result length (No result)
	if condition.Offset >= len(tmpAds) {
		return nil, nil
	}
	//Offset < Result <= Result length && Offset + Limit
	for i := condition.Offset; i < len(tmpAds) && i < condition.Offset+condition.Limit; i++ {
		var searchResult = SearchResult{
			Title: tmpAds[i].Title,
			EndAt: tmpAds[i].EndAt,
		}
		resultAds = append(resultAds, searchResult)
	}

	return resultAds, nil
}

func getAdsByCondition(condition SearchCondition) []Ad {

	//Assemble query string
	head := "SELECT UUID,title,start_at,end_at,age_start,age_end,Country,Platform,Gender FROM ad WHERE $1 BETWEEN start_at AND end_at "

	var body []string
	//Age
	if len(condition.Age) > 0 {
		ageQuery := "( "
		for index, value := range condition.Age {
			ageQuery += "( " + value + " BETWEEN age_start AND age_end)"
			if index != len(condition.Age)-1 {
				ageQuery += " OR "
			}
		}
		ageQuery += " )"
		body = append(body, ageQuery)
	}
	//Gender
	if len(condition.Gender) > 0 {
		genderQuery := "(Gender LIKE '%null%' OR "
		for index, value := range condition.Gender {
			genderQuery += "Gender LIKE '%" + value + "%'"
			if index != len(condition.Gender)-1 {
				genderQuery += " OR "
			}
		}
		genderQuery += " )"
		body = append(body, genderQuery)
	}
	//Country
	if len(condition.Country) > 0 {
		countryQuery := "(Country LIKE '%null%' OR "
		for index, value := range condition.Country {
			countryQuery += "Country LIKE '%" + value + "%'"
			if index != len(condition.Country)-1 {
				countryQuery += " OR "
			}
		}
		countryQuery += " )"
		body = append(body, countryQuery)
	}
	//Platform
	if len(condition.Platform) > 0 {
		platformQuery := "(Platform LIKE '%null%' OR "
		for index, value := range condition.Platform {
			platformQuery += "Platform LIKE '%" + value + "%'"
			if index != len(condition.Platform)-1 {
				platformQuery += " OR "
			}
		}
		platformQuery += " )"
		body = append(body, platformQuery)
	}

	tail := " ORDER BY end_at"
	if len(body) != 0 {
		head += " AND "
	}
	query := head + strings.Join(body, " AND ") + tail
	//println(query)
	rows, err := dbClient.Query(query, getNowTime())
	if err != nil {
		print(err.Error())
		return nil
	}
	defer rows.Close()

	//Mapping
	var ads = []Ad{}
	for rows.Next() {
		var countryJson string
		var platformJson string
		var genderJson string
		var ad Ad
		err := rows.Scan(&ad.UUID, &ad.Title, &ad.StartAt, &ad.EndAt, &ad.Conditions.AgeStart, &ad.Conditions.AgeEnd, &countryJson, &platformJson, &genderJson)
		if err != nil {
			println("Cannot map result ", err.Error())
			return nil
		}
		err = json.Unmarshal([]byte(countryJson), &ad.Conditions.Countries)
		err = json.Unmarshal([]byte(platformJson), &ad.Conditions.Platforms)
		err = json.Unmarshal([]byte(genderJson), &ad.Conditions.Gender)
		ads = append(ads, ad)
	}

	return ads
}

//func getAdsById(ids []string) []Ad {
//	var ads = []Ad{}
//
//	for _, id := range ids {
//		//Get ads from DB
//		query := "SELECT UUID,title,start_at,end_at,age_start,age_end,Country,Platform,Gender FROM ad WHERE ($1 BETWEEN start_at AND end_at) AND UUID=$2 ORDER BY end_at"
//		rows, err := dbClient.Query(query, getNowTime(), id)
//		if err != nil {
//			println("Cannot get ad by id: ", err.Error())
//			return nil
//		}
//		defer rows.Close()
//
//		//Mapping
//		var countryJson string
//		var platformJson string
//		var genderJson string
//		var ad Ad
//		rows.Next()
//		err = rows.Scan(&ad.UUID, &ad.Title, &ad.StartAt, &ad.EndAt, &ad.Conditions.AgeStart, &ad.Conditions.AgeEnd, &countryJson, &platformJson, &genderJson)
//		if err != nil {
//			print("Cannot map result ", err.Error())
//			return nil
//		}
//		err = json.Unmarshal([]byte(countryJson), &ad.Conditions.Countries)
//		err = json.Unmarshal([]byte(platformJson), &ad.Conditions.Platforms)
//		err = json.Unmarshal([]byte(genderJson), &ad.Conditions.Gender)
//		ads = append(ads, ad)
//	}
//
//	return ads
//}
