/** Short CSV snippets shared by CopyableCsv previews and InvokeTabs HTTP bodies. */

export const SALES_SAMPLE = `order_date,region,category,product,quantity,amount
2024-01-01,Central,Hardware,Connector,16,2488.24
2024-01-02,East,Tools,Gear,42,207.08
2024-01-03,North,Mechanical,Sensor,20,3465.22
2024-01-04,West,Electronics,Valve,24,7633.44
2024-03-12,South,Industrial,Widget,31,5088.10
2024-06-18,East,Electronics,Relay,12,2214.50
2024-09-05,North,Hardware,Bolt,28,1724.27
2024-12-20,West,Tools,Gadget,19,2938.82
2025-02-08,Central,Mechanical,Valve,27,7350.26
2025-04-14,South,Electronics,Widget,41,3278.77
2025-07-22,East,Industrial,Connector,15,5093.42
2025-10-03,North,Tools,Gear,33,4102.15
2025-11-19,West,Hardware,Sensor,22,2890.40
2025-12-28,South,Mechanical,Gadget,26,611.32`;

export const QUICK_SALES_SAMPLE = `region,product,quantity
West,Widget,39
West,Gear,20
West,Sensor,28
West,Relay,14
West,Bolt,33
South,Widget,15
South,Gear,42
South,Sensor,47
South,Relay,22
South,Bolt,18
North,Widget,28
North,Gear,43
North,Sensor,19
North,Relay,36
North,Bolt,25
East,Widget,31
East,Gear,22
East,Sensor,35
East,Relay,12
East,Bolt,40
Central,Widget,35
Central,Gear,17
Central,Sensor,19
Central,Relay,29
Central,Bolt,24`;

export const DIMENSIONS_SALES_SAMPLE = `region,product,sales
Asia,Widget,100
Asia,Gadget,80
EU,Widget,60`;

export const CHORD_RELATIONS_SAMPLE = `source,target,value
A,B,14
A,C,8
B,C,20
B,E,15
C,B,8
C,E,3
D,A,12
D,B,3
E,A,15
E,C,5
F,C,5
G,A,6
G,B,8
G,D,4`;

export const SANKEY_FLOWS_SAMPLE = `name,source,target,value,cost
web,Visit,Landing,5000,12
web,Landing,Signup,1200,8
web,Landing,Exit,3800,2
web,Signup,Trial,800,15
web,Signup,Newsletter,400,5
web,Trial,Paid,320,25
web,Trial,Churn,480,3
app,Install,Onboard,3000,10
app,Onboard,Active,1800,18
app,Onboard,Dormant,1200,4
app,Active,Subscribe,900,30
app,Active,Free,900,6
app,Subscribe,Renew,600,22
app,Subscribe,Cancel,300,5`;

export const RADAR_PROFILES_SAMPLE = `subject,metric,score,tier
Pond,throughput,92,eager
Pond,latency,70,eager
Pond,memory,80,eager
Pond,stability,88,eager
Chi,throughput,85,eager
Chi,latency,78,eager
Chi,memory,74,eager
Chi,stability,82,eager
Gin,throughput,90,lazy
Gin,latency,65,lazy
Gin,memory,88,lazy
Gin,stability,79,lazy
Echo,throughput,88,lazy
Echo,latency,72,lazy
Echo,memory,81,lazy
Echo,stability,84,lazy`;

export const REGION_METRICS_SAMPLE = `region,latency,sales
North,12.4,8200
South,18.1,6400
East,9.7,9100
West,14.2,7300
Central,11.0,7800`;
