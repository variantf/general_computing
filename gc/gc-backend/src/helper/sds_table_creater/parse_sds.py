from he_14_map import he_data_mapping as data_he14
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
		if 'type' in d and 'field' in d and d['type'] == 'normal':
			t = {}
			convert = {}
			for col_name, val in d['field'].items():
				sheet_name = val[0]
				t['name'] = sheet_name
				field_name = val[1]
				field_type = 'float'
				if field_name == 'None':
					continue
				convert[col_name] = field_name
			t['fields'] = convert
			s[xlsx_name] = t

def output(s):
	print(json.dumps(s, ensure_ascii = False, indent = 2))

if __name__ == '__main__':
	parse(data_he14)
	parse(data_he15)
	output(s)
