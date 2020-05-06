package handler

import (
	"Ihome/IhomeWeb/utils"
	"context"
	"encoding/json"
	"github.com/afocus/captcha"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
	"image/color"
	"time"

	"github.com/micro/go-log"

	example "Ihome/GetImageCd/proto/getimagecd"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) Call(ctx context.Context, req *example.Request, rsp *example.Response) error {
	beego.Info("获取验证码图片")
	cap := captcha.New()
	if err := cap.SetFont("comic.ttf"); err != nil {
		beego.Error(err)
	}
	cap.SetSize(90, 41)
	cap.SetDisturbance(captcha.NORMAL)
	cap.SetFrontColor(color.RGBA{255, 255, 255, 255})
	cap.SetBkgColor(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 153, 0, 255})
	img, str := cap.Create(4, captcha.NUM)
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
	bm.Put(req.Uuid, str, time.Minute*5)
	rsp.Error = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Error)
	rsp.Pix = img.Pix
	rsp.Stride = int64(img.Stride)
	rsp.Max = &example.Response_Point{X: int64(img.Rect.Max.X), Y: int64(img.Rect.Max.Y)}
	rsp.Min = &example.Response_Point{X: int64(img.Rect.Min.X), Y: int64(img.Rect.Min.Y)}
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
