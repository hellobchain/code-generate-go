package e

import "github.com/pkg/errors"

type ErrCode int

func (e ErrCode) Error() string {
	return GetMsg(e)
}

func GetMsg(e ErrCode) string {
	m, ok := msgFlags[e]
	if !ok {
		return "未知错误"
	}
	return m[0]
}

func NewError(code ErrCode, err error) error {
	if code == SUCCESS {
		return nil
	}
	if err != nil {
		return errors.Wrapf(code, "%v", err)
	}
	return code
}
