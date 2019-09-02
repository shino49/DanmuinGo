package Danmu

import(
    "os"
    "fmt"
    "log"
    "encoding/json"
    "golang.org/x/net/websocket"
    "io/ioutil"
    "sync"
)

type Barrage struct{
    Value string `json:"value"`
    Time float64 `json:"time"`
    Color string `json:"color"`
    Fontsize int `json:"fontzize"`
    speed float64 `json:"speed"`
    opacity float64`json:"opacity"`
}

type BarrageDataInit struct{
    Typed string `json:"typed"`
    Data []Barrage `json:"data"`
}

type BarrageDataAdd struct{
    Typed string `json:"typed"`
    Data Barrage `json:"data"`
}

//var danmuPool []Barrage = []Barrage{{Value:"test1",Time:2.0,Color:"red",Fontsize:20,speed:1.0,opacity:0.7},{Value:"test2",Time:2.0,Color:"red",Fontsize:20,speed:1.0,opacity:0.7}}

type DanmuPool []Barrage

type DanmuCargoAll struct{
    sync.Mutex
    DanmuCargo map[string] DanmuPool `json:"danmuCargo"`
    UserPool map[string] int64 `json:"userPool"`
}

func (this *DanmuCargoAll) PoolInit() {
    this.DanmuCargo = make(map[string] DanmuPool)
    this.UserPool = make(map[string] int64)
    fileInfo, err := os.Stat("danmucargo.json")
    fmt.Println(fileInfo)
    if err != nil {
        if os.IsNotExist(err) {
            //log.Fatal("File does not exist.")
            jsonFile, err2 := os.OpenFile("danmucargo.json",os.O_WRONLY|os.O_TRUNC|os.O_CREATE,0666)
            if err2 != nil {
                log.Fatal(err2)
            }
            defer jsonFile.Close()
            jsonByte,_ := json.Marshal(this.DanmuCargo)
            _, err3 := jsonFile.Write(jsonByte)
            if err3 != nil {
                log.Fatal(err3)
            }
            log.Println("create new json!")

        }
    }else{
        jsonData, err := ioutil.ReadFile("danmucargo.json")
        if err != nil {
            log.Fatal(err)
        }
        json.Unmarshal(jsonData, this.DanmuCargo)

    }
}

func (this *DanmuCargoAll) save() {
    jsonFile, err := os.OpenFile("danmucargo.json",os.O_WRONLY|os.O_TRUNC|os.O_CREATE,0666)
    if err != nil {
        log.Fatal(err)
    }
    jsonByte,jsonerr:= json.Marshal(this.DanmuCargo)
    _, err2 := jsonFile.Write(jsonByte)
    if err2 != nil {
                log.Fatal(err2)
    }
    log.Println("save as json success")
    //log.Println(string(jsonByte))
    log.Println(jsonerr)
    log.Println("jsonsave")
}

func DanmuRec(ws *websocket.Conn,vname string){
    var jsonRec string
    var danmu Barrage
    for{
         if err := websocket.Message.Receive(ws, &jsonRec); err != nil {
            log.Println("Can't receive")
            break
        }
        log.Println("Received danmu from client: " + jsonRec)
        json.Unmarshal([]byte(jsonRec), &danmu)
        //EdanmuCargoAll.RLock()
        EdanmuCargoAll.Lock()
        if EdanmuCargoAll.DanmuCargo[vname][0].Value == ""{
            EdanmuCargoAll.DanmuCargo[vname][0]=danmu
            jsonSend, _ := json.Marshal(BarrageDataAdd{Typed:"add",Data:danmu})
            jsonSendStr := string(jsonSend)
            log.Println("send danmu : ",jsonSendStr)
            if err2 := websocket.Message.Send(ws, jsonSendStr); err2 != nil {
                fmt.Println("Can't send")
                break
            }

        }else{
            EdanmuCargoAll.DanmuCargo[vname]  = append(EdanmuCargoAll.DanmuCargo[vname] ,danmu)
        }
        EdanmuCargoAll.Unlock()
        //EdanmuCargoAll.UnRLock()
        //EdanmuCargoAll.DanmuCargo[vname] = danmuPool
        EdanmuCargoAll.save()
        //var danmuAdd BarrageDataAdd = BarrageDataAdd{Typed:"add",Barrage:danmu}
        //jsonSend, err2 := json.Marshal(danmuAdd)
    }
}
func DanmuWs(ws *websocket.Conn) {
    log.Println("ws connected success in /danmuWs/")
    var vname string 
    if err := websocket.Message.Receive(ws, &vname); err != nil {
        log.Println("Can't receive")
        return
    }
    if vname == ""{
        log.Fatal("error,no vname")
        return
    }
    //EdanmuCargoAll.RLock()
    EdanmuCargoAll.Lock()
    if danmuPool,ok := EdanmuCargoAll.DanmuCargo[vname]; !ok{
        EdanmuCargoAll.DanmuCargo[vname] = make(DanmuPool,1)
        log.Println(danmuPool)

        //danmuPool = EdanmuCargoAll.DanmuCargo[vname]
    }
    var initNum = len(EdanmuCargoAll.DanmuCargo[vname])
    EdanmuCargoAll.Unlock()
    var barrangeinit BarrageDataInit = BarrageDataInit{Typed:"init",Data:EdanmuCargoAll.DanmuCargo[vname] }
    //EdanmuCargoAll.RUnlock()
    jsonSendInit, _ := json.Marshal(barrangeinit)//unmershal输出[]byte型，需转换string型
    jsonSendInitStr := string(jsonSendInit)
    log.Println("send danmu: "+jsonSendInitStr)
    if err2 := websocket.Message.Send(ws, jsonSendInitStr); err2 != nil {
                fmt.Println("Can't send")
                return
            }
    go DanmuRec(ws,vname)
    for {
        EdanmuCargoAll.Lock()
        if initNum < len(EdanmuCargoAll.DanmuCargo[vname] ){
            var danmuAdd BarrageDataAdd = BarrageDataAdd{Typed:"add",Data:EdanmuCargoAll.DanmuCargo[vname] [len(EdanmuCargoAll.DanmuCargo[vname])-1]}
            EdanmuCargoAll.Unlock()
            initNum++
            jsonSend, _ := json.Marshal(danmuAdd)
            jsonSendStr := string(jsonSend)
            log.Println("send danmu : ",jsonSendStr)
            if err3 := websocket.Message.Send(ws, jsonSendStr); err3 != nil {
                fmt.Println("Can't send")
                break
            }
        }else{
            EdanmuCargoAll.Unlock()
        }
    }
}

var EdanmuCargoAll DanmuCargoAll