from he_15_map import he_data_mapping as data_he15
import json

s = {}
types = {
	'float': 0,
	'string': 1,
	'boolean': 2
}

def parse(data):
	for xlsx_name, d in data.items():
		if xlsx_name == 'SB_SDS_JMCZ_14ND_ZCZJTXQKB' and 'type' in d and 'field' in d and d['type'] == '2d':
			t = {}
			convert = {}
			for col_name, val in d['field'].items():
				sheet_name = val[0]
				t['name'] = sheet_name
				field_name = val[1]
				field_type = 'float'
				if field_name == 'None':
					continue
				convert[col_name[1] + col_name[0]] = field_name
			t['fields'] = convert
			s[xlsx_name] = t

def output(s):
	print(json.dumps(s, ensure_ascii = False, indent = 2))

if __name__ == '__main__':
	parse(data_he15)
	output(s)
