package kdp

type PipeItemRoute struct {
	RouteItemID    string      `key:"1" label:"Route ItemID"`
	ConditionField string      `label:"Field"`
	ConditionValue interface{} `label:"Value" kf-control:"text"`
}

type PipeItem struct {
	ID             string          `kf-pos:"1,1" key:"1" label:"Name" required:"true"`
	WorkerID       string          `kf-pos:"1,2" label:"Worker" required:"true"`
	CollectProcess bool            `kf-pos:"2,1" label:"Collect Process"`
	CloseWhenDone  bool            `kf-pos:"3,1" label:"Close if done"`
	CloseWhenFail  bool            `kf-pos:"3,2" label:"Close if fail"`
	Routes         []PipeItemRoute `kf-pos:"4,1" label:"Route"`
}
