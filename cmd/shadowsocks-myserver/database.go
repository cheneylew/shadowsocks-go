package main

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/cheneylew/goutil/utils"
	"fmt"
	"time"
)

var o orm.Ormer

func init() {
	url := dbUrl("cheneylew","12344321","47.91.151.207","3308","shadowsocks-servers")
	utils.JJKPrintln(url)
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", url)
}

func dbUrl(user, password, host, port, dbName string) string {
	return fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8`, user, password, host, port, dbName)
}

func dbQueryServers(ip string) []*Server {
	var objects []*Server

	qs := o.QueryTable("server")
	_, err := qs.Filter("ip", ip).All(&objects)

	if err != nil {
		return objects
	}

	return objects
}

func dbQueryPortsWithSid(sid int64) []*Port {
	var objects []*Port

	qs := o.QueryTable("port")
	_, err := qs.Filter("server__server_id", sid).All(&objects)

	if err != nil {
		return objects
	}

	return objects
}

func dbQueryPortsWithIP(ip string) []*Port {
	var objects []*Port

	qs := o.QueryTable("port")
	_, err := qs.Filter("server__ip", ip).All(&objects)

	if err != nil {
		return objects
	}

	return objects
}

func dbQueryUsersWithUid(uid int64) []*User {
	var objects []*User
	qs := o.QueryTable("user")
	_, err := qs.Filter("user_id", uid).All(&objects)

	if err != nil {
		return objects
	}

	return objects
}

func dbQueryMyListenPorts() []*Port {
	curIp, _ := utils.ExtranetIP()
	ports := dbQueryPortsWithIP(curIp)

	var filterPorts []*Port
	for _, port := range ports {
		if port.Ptype == 0 {
			//包年包月
			//判断截止时间
			if time.Now().Before(port.End_time) {
				filterPorts = append(filterPorts, port)
			}
		} else if port.Ptype == 1 {
			//包流量
			//流量是否超限
			if port.Flow_total < port.Flow_in_max {
				filterPorts = append(filterPorts, port)
			}
		}
	}


	return filterPorts
}

func dbStart()  {
	o = orm.NewOrm()
	o.Using("default") // 默认使用 default，你可以指定为其他数据库

}

func dbUploadFlow()  {
	utils.JJKPrintln("upload ....")

	ports := dbQueryMyListenPorts()
	for _, port := range ports {
		flowCounter, ok := SS_FlowCounterManager.get(port.Port)
		if ok {
			if flowCounter.In != 0 || flowCounter.Out != 0 {
				port.Flow_in += float64(flowCounter.In)
				port.Flow_out += float64(flowCounter.Out)
				port.Flow_total += float64(flowCounter.In + flowCounter.Out)

				port.Sync_time = time.Now().Add(time.Hour*8)
				n, e := o.Update(port,"Flow_in","Flow_out","Flow_total","Sync_time")
				if n > 0 {
					//已上传 清零
					SS_FlowCounterManager.update(0,0,port.Port)
				} else {
					utils.JJKPrintln("update flow failed!!",e)
				}
			}
		}
	}
}