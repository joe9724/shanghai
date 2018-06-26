package main

import (
	"net/http"
	_ "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"
	"io/ioutil"
	"github.com/json-iterator/go"
	xj "github.com/basgys/goxml2json"
	"strings"
	"shanghai/utils"
	"shanghai/models"
)

var jsonf = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	//依次执行
	//抓取xml线路
	//CrawLineNameList()
	//按照线路获取line_id等信息,同时插入updown表
	//CrawLineIdUpdownList()
	//获取线路站台
    //CrawLineStation()
    //获取线路信息
    CrawLineInfo()

}

func CrawLineInfo() {
	db, err := utils.OpenConnection()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	var lines []models.DbLineModel
	db.Raw("select line_id,line_name from btk_lines").Find(&lines)
	for i:=0; i<len(lines);i++  {
		resp, err := http.Get("http://xxbs.sh.gov.cn:8080/weixinpage/HandlerBus.ashx?action=One&name="+lines[i].LineName)
		if err != nil {
			// handle error
		}

		// defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("err is",err.Error())
		}else{
			fmt.Println("resp is",string(body))
			var lineinfo models.UpdownModel
			err1 := jsonf.Unmarshal(body, &lineinfo)
			if err1!= nil {
				fmt.Println("err1.err is",err1.Error())
			}else{
				db.Exec("update btk_lines set end_earlytime=?,end_latetime=?,end_stop=?,start_earlytime=?,start_latetime=?,start_stop=? where line_id=?",lineinfo.EndEarlytime,lineinfo.EndLatetime,lineinfo.EndStop,lineinfo.StartEarlytime,lineinfo.StartLatetime,lineinfo.StartStop,lines[i].LineId)
				fmt.Println("update btk_lines set end_earlytime=?,end_latetime=?,end_stop=?,start_earlytime=?,start_latetime=?,start_stop=? where line_id=?",lineinfo.EndEarlytime,lineinfo.EndLatetime,lineinfo.EndStop,lineinfo.StartEarlytime,lineinfo.StartLatetime,lineinfo.StartStop,lines[i].LineId)
			}
		}
	}
}

func CrawLineStation() {
	db, err := utils.OpenConnection()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	var lines []models.DbLineModel
	db.Raw("select line_id,line_name from btk_lines").Find(&lines)
	for i:=0; i<len(lines);i++  {
		resp, err := http.Get("http://xxbs.sh.gov.cn:8080/weixinpage/HandlerBus.ashx?action=Two&lineid="+lines[i].LineId+"&name="+lines[i].LineName)
		if err != nil {
			// handle error
		}

		// defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("err is",err.Error())
		}else{
			fmt.Println("resp is",string(body))
			var lsm models.LineStationModel
			err1 := jsonf.Unmarshal(body, &lsm)
			if err1!= nil {
				fmt.Println("err1.err is","http://xxbs.sh.gov.cn:8080/weixinpage/HandlerBus.ashx?action=Two&lineid="+lines[i].LineId+"&name="+lines[i].LineName,err1.Error())
			}else{
				//往line_station插入数据
				fmt.Println("lsm.0 is",lsm.LineResults0.Direction)
				fmt.Println("lsm.1 is",lsm.LineResults1.Direction)
				fmt.Println("lsm.LineResults0.Stops.len is",len(lsm.LineResults0.Stops))
				if lsm.LineResults0.Direction == "true"{
					for k:=0;k<len(lsm.LineResults0.Stops) ;k++  {
						fmt.Println("insert into btk_line_station(line_id,line_name,line_updown,st_name,stop_id) values(?,?,?,?,?)",lines[i].LineId,lines[i].LineName,0,lsm.LineResults0.Stops[k].Zdmc,lsm.LineResults0.Stops[k].Id)
						db.Exec("insert into btk_line_station(line_id,line_name,line_updown,st_name,stop_id) values(?,?,?,?,?)",lines[i].LineId,lines[i].LineName,0,lsm.LineResults0.Stops[k].Zdmc,lsm.LineResults0.Stops[k].Id)
					}
				}
				if lsm.LineResults1.Direction == "false"{
					for k:=0;k<len(lsm.LineResults1.Stops) ;k++  {
						db.Exec("insert into btk_line_station(line_id,line_name,line_updown,st_name,stop_id) values(?,?,?,?,?)",lines[i].LineId,lines[i].LineName,1,lsm.LineResults1.Stops[k].Zdmc,lsm.LineResults1.Stops[k].Id)
					}
				}
			}
		}
	}
}

func CrawLineIdUpdownList() {
	fmt.Println("抓取开始")
	db, err := utils.OpenConnection()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	//http://61.132.47.90:8998/BusService/Require_AllRouteData/?TimeStamp=123
	//先获取所有线路
	resp, err := http.Get("http://www.gembo.cn/app/shbus/all2018.xml")
	if err != nil {
		// handle error
	}

	// defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	xml := strings.NewReader(string(body))
	json, err := xj.Convert(xml)
	if err != nil {
		panic("That's embarrassing...")
	}

	fmt.Println(json.String())
	var clm models.CrawLineModel

	err1 := jsonf.Unmarshal([]byte(json.String()), &clm)
	if err1 != nil {
		fmt.Println("err1 is", err.Error())
	} else {
		for i := 0; i < len(clm.Lines.Line); i++ {
			resp, err := http.Get("http://xxbs.sh.gov.cn:8080/weixinpage/HandlerBus.ashx?action=One&name="+clm.Lines.Line[i].Name)
			if err != nil {
				// handle error
			}

			// defer resp.Body.Close()
			bb, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				// handle error
			}
			var updown models.UpdownModel

			err3 := jsonf.Unmarshal(bb, &updown)
			if err3 !=nil {
				fmt.Println("err3.err is",clm.Lines.Line[i].Name,string(bb),err3.Error())
			}else{
				db.Exec("update btk_lines set line_id=? where line_name=?",updown.LineId,clm.Lines.Line[i].Name)
			}
			//if i==0{
			//db.Exec("insert into btk_lines(line_id,line_name) values(?,?)", clm.Lines.Line[i].Value, clm.Lines.Line[i].Name)
			//fmt.Println("insert into btk_lines(line_id,line_name) values(?,?)", clm.Lines.Line[i].Value, clm.Lines.Line[i].Name)
			//}

		}
	}

	fmt.Println("抓取完成!")
}

func CrawLineNameList() {
	fmt.Println("抓取开始")
	db, err := utils.OpenConnection()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	//http://61.132.47.90:8998/BusService/Require_AllRouteData/?TimeStamp=123
	//先获取所有线路
	resp, err := http.Get("http://www.gembo.cn/app/shbus/all2018.xml")
	if err != nil {
		// handle error
	}

	// defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	xml := strings.NewReader(string(body))
	json, err := xj.Convert(xml)
	if err != nil {
		panic("That's embarrassing...")
	}

	fmt.Println(json.String())
	var clm models.CrawLineModel

	err1 := jsonf.Unmarshal([]byte(json.String()), &clm)
	if err1 != nil {
		fmt.Println("err1 is", err.Error())
	} else {
		for i := 0; i < len(clm.Lines.Line); i++ {
			//if i==0{
			db.Exec("insert into btk_lines(line_id,line_name) values(?,?)", clm.Lines.Line[i].Value, clm.Lines.Line[i].Name)
			fmt.Println("insert into btk_lines(line_id,line_name) values(?,?)", clm.Lines.Line[i].Value, clm.Lines.Line[i].Name)
			//}

		}
	}

	fmt.Println("抓取完成!")

}
