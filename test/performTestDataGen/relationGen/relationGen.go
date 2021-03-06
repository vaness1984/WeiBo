/*
@Author : Ryan.wuxiaoyong
*/

package main


import (
	//"WeiBo/common"
	"context"
	//"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
	"time"

	//"time"
	//"sync/atomic"
	"encoding/json"
	"io/ioutil"
	//"sync"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//配置文件格式
type configType struct {
	MongoDBUrl string `json: "mongoDBUrl"`
	MongoDBName string `json: "mongoDBName"`
}
//配置文件数据对象
var gConfig = &configType{}

func main(){
	if len(os.Args)!=2 {
		log.Fatalf("xxx configPath")
	}
	confPath := os.Args[1]

	//
	confBytes, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalf("Read config file failed.[%s][%+v]", confPath, err)
	}
	//解析
	err = json.Unmarshal(confBytes, gConfig)
	if err != nil {
		log.Fatalf("Read config file failed.[%s][%+v]", confPath, err)
	}

	mongoOpt := options.Client().ApplyURI("mongodb://" + gConfig.MongoDBUrl)
	//if usrName != "" {
	//	mongoOpt = mongoOpt.SetAuth(options.Credential{Username:usrName, Password:pass})
	//}
	//超时设置
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	client, err := mongo.Connect(ctx, mongoOpt)
	cancel()
	if err != nil {
		log.Fatalf("mongodb connect failed. [%+v]", err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, nil)
	cancel()
	if err != nil {
		log.Fatalf("mongodb ping failed. [%+v]", err)
	}

	log.Printf("init")


	for i := range gUserFollower{
		gUserFollower[i].followerId = make(map[int64]bool)
	}
	for i := range gUserFollow{
		gUserFollow[i].followId = make(map[int64]bool)
	}

	//用户id（10000000~19999999）一千万用户，每人10个粉丝。
	createFollowerX(10000000, 10000000, 10)
	//用户id（20000000~20009999）一万用户，每人一千粉丝
	createFollowerX(20000000, 10000, 1000)
	//用户id（20010000~20007999）八千用户，每人三千粉丝
	createFollowerX(20010000, 8000, 3000)
	//用户id（20018000~20019999）两千用户，每人一万粉丝
	createFollowerX(20018000, 2000, 10000)
	//用户id（20020000~20020999）一千用户，每人两万粉丝
	createFollowerX(20020000, 1000, 20000)
	//用户id（20021000~20021799）八百用户，每人三万粉丝
	createFollowerX(20021000, 800, 30000)
	//用户id（20021800~20019999）两百用户，每人十万粉丝
	createFollowerX(20021800, 200, 100000)
	//用户id（20022000~20022099）一百用户，每人二十万粉丝
	createFollowerX(20022000, 100, 200000)
	//用户id（20022100~20022199）一百用户，每人三十万粉丝
	createFollowerX(20022100, 100, 300000)
	//用户id（20022200~20022299）一百用户，每人四十万粉丝
	createFollowerX(20022200, 100, 400000)
	//用户id（20022300~20022302）三个用户，每人100万粉丝
	createFollowerX(20022300, 3, 1000000)
	//用户id（20022303~20022305）三个用户，每人150万粉丝
	createFollowerX(20022303, 3, 1500000)
	//用户id（20022306~20022308）三个用户，每人200万粉丝
	createFollowerX(20022306, 3, 2000000)
	//用户id（20022309~20022309）一个用户，每人300万粉丝
	createFollowerX(20022309, 1, 3000000)



	storeUserFollowerInDB(client)
	//不用了，清掉
	gUserFollower = [10022310]userFollowerT{}

	storeUserFollowInDB(client)
	//不用了，清掉
	gUserFollow = [10000000]userFollowT{}


}

//粉丝表
type userFollowerT struct {
	userId int64
	followerId map[int64]bool
}
var gUserFollower = [10022310]userFollowerT{}

//用户关注表
type userFollowT struct {
	userId int64
	followId map[int64]bool
}
var gUserFollow = [10000000]userFollowT{}

func createFollowerX(userStartId int64, userCount int, followerCount int){
	for i:=0; i<userCount; i++{
		if userCount >=10 && i%(userCount/10)==0{
			log.Printf("createFollowerX(userStartId=%d, userCount=%d, followerCount=%d) finished: %d/%d",
				userStartId, userCount, followerCount, i, userCount)
		}

		userId := int64(userStartId + int64(i))
		for x:=0; x<followerCount; x++{
			//随机一个粉丝id
			followerId := rand.Int63n(10000000) + 10000000

			//粉丝不要重复
			{
				_, ok := gUserFollower[userId-10000000].followerId[followerId]
				if ok{
					//重复了
					x--
					continue
				}
			}
			//关注不要重复
			{
				_, ok := gUserFollow[followerId-10000000].followId[userId]
				if ok{
					//重复了
					x--
					continue
				}
			}

			gUserFollower[userId-10000000].userId = userId
			gUserFollower[userId-10000000].followerId[followerId] = true
			gUserFollow[followerId-10000000].userId = followerId
			gUserFollow[followerId-10000000].followId[userId] = true
		}
	}

	log.Printf("createFollowerX(userStartId=%d, userCount=%d, followerCount=%d) finished.",
		userStartId, userCount, followerCount)
}

func storeUserFollowInDB(client *mongo.Client){
	log.Printf("storeUserFollowInDB start")

	curIndex := 0
	for {
		log.Printf("progress %d/%d", curIndex, len(gUserFollow))

		count := 0

		now := time.Now().Unix()
		var dataArr []interface{}
		for ; curIndex < len(gUserFollow) && count < 100000; curIndex++ {
			count++
			info := gUserFollow[curIndex]
			for followId := range info.followId {
				data := bson.D{{"userid", info.userId}, {"followid", followId}, {"followtime", now}}
				dataArr = append(dataArr, data)
			}
		}

		colName := fmt.Sprintf("Follow")
		collection := client.Database(gConfig.MongoDBName).Collection(colName)
		_, err := collection.InsertMany(context.TODO(), dataArr)
		if err != nil {
			log.Fatalf("InsertMany failed. err=[%+v]", err)
		}

		if curIndex == len(gUserFollow){
			break
		}
	}

	log.Printf("storeUserFollowInDB finished.")
}
func storeUserFollowerInDB(client *mongo.Client){
	log.Printf("storeUserFollowerInDB start")

	curIndex := 0
	for {
		log.Printf("progress %d/%d", curIndex, len(gUserFollower))

		count := 0

		now := time.Now().Unix()
		//
		var dataArr []interface{}
		for ; curIndex < len(gUserFollower); curIndex++ {
			if curIndex < 10000000{
				if count >= 5000{
					break
				}
			}else{
				if count >= 1{
					break
				}
			}
			count++

			info := gUserFollower[curIndex]
			for followerId := range info.followerId {
				data := bson.D{{"userid", info.userId}, {"followedid", followerId}, {"followtime", now}}
				dataArr = append(dataArr, data)
			}
		}
		colName := fmt.Sprintf("Follower")
		collection := client.Database(gConfig.MongoDBName).Collection(colName)
		_, err := collection.InsertMany(context.TODO(), dataArr)
		if err != nil {
			log.Fatalf("InsertMany failed. err=[%+v]", err)
		}

		if curIndex == len(gUserFollower){
			break
		}
	}

	log.Printf("storeUserFollowInDB finished.")
}

