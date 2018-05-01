package sqlx

import (
	"github.com/sirupsen/logrus"
)

var l = logrus.New()

func init() {
	l.SetLevel(logrus.DebugLevel)
}
