package ssh

import (
	"github.com/Superdanda/hade/framework/provider/config"
	"github.com/Superdanda/hade/framework/provider/log"
	tests "github.com/Superdanda/hade/test"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHadeSSHService_Load(t *testing.T) {
	container := tests.InitBaseContainer()
	container.Bind(&config.HadeConfigProvider{})
	container.Bind(&log.HadeLogServiceProvider{})

	Convey("test get client", t, func() {
		hadeRedis, err := NewHadeSSH(container)
		So(err, ShouldBeNil)
		service, ok := hadeRedis.(*HadeSSH)
		So(ok, ShouldBeTrue)
		client, err := service.GetClient(WithConfigPath("ssh.web-01"))
		So(err, ShouldBeNil)
		So(client, ShouldNotBeNil)
		session, err := client.NewSession()
		So(err, ShouldBeNil)
		out, err := session.Output("pwd")
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		session.Close()
	})
}
