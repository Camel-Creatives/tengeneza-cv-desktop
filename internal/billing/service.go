// Package billing owns the public pricing catalog and wallet grant rules.
package billing

type Plan struct { Code, Label string; AmountTZS int64; ExportCredits int; AITokenGrant int64; VoiceSeconds int; Subscription bool }
var plans = []Plan{{"free","Free",0,0,20000,0,false},{"starter_5","Starter Pack",3000,5,100000,0,false},{"growth_15","Growth Pack",7000,15,350000,0,false},{"pro_40","Pro Pack",15000,40,900000,3600,false},{"career_pro_monthly","Career Pro",12000,0,0,0,true},{"agency_100seats","Agency / Institution",250000,0,0,0,true}}

func Plans()[]Plan{return append([]Plan(nil),plans...)}
func PlanByCode(code string)(Plan,bool){for _,plan:=range plans{if plan.Code==code{return plan,true}};return Plan{},false}
