import data_manager_pb2 as dm
import general_computing_pb2 as gc
import formula_pb2 as formula
from datetime import datetime as datetime
import json
import grpc
import csv

#names = json.load(open('keys.json'))

channel = grpc.insecure_channel('192.168.44.199:12100')
stub = dm.DataManagerStub(channel)

# table = {
#     'path': '/风控/月度/国税',
#     'fields': {
#         'XQYKJZZ_ZCFZB:'+j+"_"+str(i): gc.Table.Field(name='XQYKJZZ_ZCFZB:'+j+"_"+str(i), type='FLOAT')
#         for i in range(1, 32) for j in ['QMYE_ZC', 'NCYE_ZC', 'QMYE_QY', 'NCYE_QY']
#     }
# }
# table['fields']['Company'] = gc.Table.Field(name="Company", type='STRING')
# table['fields']['Year'] = gc.Table.Field(name="Year", type='FLOAT')
# table['pks'] = ['Company', 'Year', 'Month']

# tb = dm.UpdateTableRequest(table=table)
# stub.UpdateTable(tb)


filename = '/Users/jiayu/SB_CWBB_XQYKJZZ_LRB.csv'
reader = csv.reader(open(filename, newline=''))
K = ['BYJE', 'BNLJJE']
K_idx = [1,2]
heads = next(reader)
print(heads)
for i in range(len(heads)):
    print(heads[i])
    for k in range(len(K)):
        if K[k] == heads[i]:
            K_idx[k] = i
    if heads[i] == 'EWBHXH':
        head_idx = i
    if heads[i] == 'NSRMC':
        company_idx = i
    if heads[i] == 'XGRQ':
        date_idx = i
print(K_idx, head_idx, date_idx)
entities = []
for row in reader:
    for i in range(len(K)):
        try:
            val = float(row[K_idx[i]])
            entity = {
                'path': '/风控/月度/国税',
                'fields': []
            }
            date = datetime.strptime(row[date_idx].replace(' ', ''), '%d-%m月-%y')
            entity['fields'].append(gc.Entity.Field(name="XQYKJZZ_LRB:"+K[i]+"_"+row[head_idx], value=formula.Literal(float_value=val)))
            #date = datetime.fromtimestamp((row[date_idx] - 25569.00) * 24 * 60 * 60)
            entity['fields'].append(gc.Entity.Field(name='Company', value=formula.Literal(string_value=row[company_idx])))
            entity['fields'].append(gc.Entity.Field(name='Year', value=formula.Literal(float_value=date.year)))
            entity['fields'].append(gc.Entity.Field(name='Month', value=formula.Literal(float_value=date.month)))
            entities.append(entity)
        except Exception as e:
            print(e)
            pass
    if len(entities) > 100000:
        print('start insert')
        stub.UpdateData(dm.UpdateDataRequest(entity=entity))
        entities = []
print('start insert')
print(stub.UpdateData(dm.UpdateDataRequest(entity=entity)))
