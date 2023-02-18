package pushc

import (
  "context"
  "errors"
  "fmt"
  "github.com/xpwu/go-log/log"
  "github.com/xpwu/go-stream/push/core"
  "strings"
  "time"
)

const (
  dataSubProtocolId = 0
  closeSubProtocol  = 1
)

func subProtocolText(sub byte) string {
  switch sub {
  case dataSubProtocolId:
    return "dataSubProtocolId"
  case closeSubProtocol:
    return "closeSubProtocol"
  default:
    return ""
  }
}

// host:port/pushToken  unix方式的host,可能也存在'/'符号
func send(ctx context.Context, pushUrl string, sub byte, data []byte, timeout time.Duration) (err error) {
  ctx, logger := log.WithCtx(ctx)

  index := strings.LastIndex(pushUrl, "/")

  if index == -1 {
    err = fmt.Errorf("pushurl(%s) error, must be '<host>/pushToken' ", pushUrl)
    logger.Error(err)
    return
  }

  url := pushUrl[:index]
  token := pushUrl[index+1:]

  if len(token) != core.TokenLen {
    err = fmt.Errorf("len token(%s) error", token)
    logger.Error(err)
    return
  }

  logger.Debug("token:" + token)

  res, err := SendTo(ctx, url, data, token, sub, timeout)
  if err != nil {
    return err
  }

  if res.State != core.Success {
    err = errors.New(core.StateText(res.State))
    logger.Error(err)
    return
  }

  return nil
}

func Close(ctx context.Context, pushUrl string) error {
  ctx, logger := log.WithCtx(ctx)

  logger.PushPrefix(fmt.Sprintf("push to %s for close that connection, ", pushUrl))

  logger.Debug("start. ")
  err := send(ctx, pushUrl, closeSubProtocol, make([]byte, 0), 30*time.Second)
  if err != nil {
    logger.Error("error, ", err)
    return err
  }
  logger.Debug("end. ")
  return nil
}

func PushData(ctx context.Context, pushUrl string, data []byte, timeout time.Duration) error {
  ctx, logger := log.WithCtx(ctx)

  logger.PushPrefix(fmt.Sprintf("push data(len=%d) to %s,", len(data), pushUrl))
  // 控制大小  默认1M
  if len(data) > 1*1024*1024 {
    log.Error("data's length must be less than 1MB")
    return errors.New("data's length must be less than 1MB")
  }

  logger.Info("start. ")
  err := send(ctx, pushUrl, dataSubProtocolId, data, timeout)
  if err != nil {
    logger.Error("error, ", err)
    return err
  }
  logger.Info("end. ")
  return nil
}
