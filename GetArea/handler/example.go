package handler

import (
	example "Ihome/GetArea/proto/getarea"
	"Ihome/IhomeWeb/models"
	"Ihome/IhomeWeb/utils"
	"context"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/orm"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
	"github.com/micro/go-log"
	"time"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetArea(ctx context.Context, req *example.Request, rsp *example.Response) error {
	rsp.Error = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Error)
	redisConf := map[string]string{
		"key":   utils.G_server_name,
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNUm": utils.G_redis_dbnum,
	}
	redisConfJs, _ := json.Marshal(redisConf)
	bm, err := cache.NewCache("redis", string(redisConfJs))
	if err != nil {
		beego.Info("redis连接失败", err)

		rsp.Error = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Error)
		return nil
	}
	areaInfo := bm.Get("area_info")

	if areaInfo != nil {
		beego.Info("获取到了地域信息:", string(areaInfo.([]byte)))
		areas := []models.Area{}
		errJson := json.Unmarshal(areaInfo.([]byte), &areas)
		if errJson != nil {
			beego.Info("数据缓存失败:", errJson)
			rsp.Error = utils.RECODE_DATAERR
			rsp.Errmsg = utils.RecodeText(rsp.Error)
			return nil
		}
		beego.Info("得到从缓存中提取的area数据,", areas)
		for _, area := range areas {
			tmp := example.Response_Areas{Aid: int32(area.Id), Aname: area.Name}
			rsp.Data = append(rsp.Data, &tmp)
		}
		return nil
	}

	o := orm.NewOrm()
	qs := o.QueryTable("area")
	var areas []models.Area
	num, err := qs.All(&areas)
	if err != nil {
		beego.Info("数据库查询失败:", err)
		rsp.Error = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Error)
		return nil
	}
	if num == 0 {
		beego.Info("数据库没有数据")
		rsp.Error = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Error)
		return nil
	}
	areaJson, _ := json.Marshal(areas)
	err = bm.Put("area_info", areaJson, time.Hour*1)
	if err != nil {
		beego.Info("数据缓存失败:", err)
		rsp.Error = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Error)
		return nil
	}
	for _, area := range areas {
		tmp := example.Response_Areas{Aid: int32(area.Id), Aname: area.Name}
		rsp.Data = append(rsp.Data, &tmp)
	}
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Example) Stream(ctx context.Context, req *example.StreamingRequest, stream example.Example_StreamStream) error {
	log.Logf("Received Example.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Logf("Responding: %d", i)
		if err := stream.Send(&example.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Example) PingPong(ctx context.Context, stream example.Example_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Logf("Got ping %v", req.Stroke)
		if err := stream.Send(&example.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
