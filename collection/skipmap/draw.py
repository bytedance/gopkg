import os
import sys
# from pyecharts.charts import Bar,Line
# from pyecharts import options as opts

# python3 visul.py benchamrk.txt (raw)
# Need `benchstat`(go) and `pyecharts`(pip)
target = sys.argv[1]
target_type = "raw"
if len(sys.argv) >= 3:
    target_type = sys.argv[2]

if target_type == "raw":
    x = os.popen('benchstat ' + target).read()
    infos = x.split("\n")[1:]
else:
    with open(target) as f:
        infos = f.read().split("\n")[1:]

results = []

for i in infos:
    if i == "":
        continue
    if "alloc/op" in i:
        break
    prefix = ""
    for idx, q in enumerate(i):
        if q == " ":
            prefix = i[:idx].replace(" ","")
            suffix = i[idx:]
            break
    if prefix == "":
        raise Exception(i)

    p = prefix.split("/")
    benchfunc = p[0]
    benchfrom = p[1].split("-")[0]
    # benchcore = p[1].split("-")[1]
    benchspeed = suffix.replace(" ", "").split("±")[0]
    if benchspeed.find("ns") != -1:
        benchspeed = float(benchspeed.replace("ns",""))
    elif benchspeed.find("µs") != -1:
        benchspeed = float(benchspeed.replace("µs","")) * 1000
    elif benchspeed.find("ms") != -1:
        benchspeed = float(benchspeed.replace("ms","")) * 1000000
    results.append([benchfunc,benchfrom,benchspeed])

first = results[0][1]
second = results[1][1]

xaxis = []
for i in results:
    if i[0] not in xaxis:
        xaxis.append(i[0])


# bar = (
#     Line()
#         .add_xaxis(xaxis)
#         .add_yaxis(first, [i[2] for i in results if i[1] == first],is_symbol_show=False,symbol="diamond",symbol_size=30)
#         .add_yaxis(second, [i[2] for i in results if i[1] == second],is_symbol_show=False,symbol="diamond",symbol_size=30)
#         .set_global_opts(title_opts=opts.TitleOpts(title="benchmark, lower is better"))
# )
print(" ".join(xaxis))
print( first," ".join([str(i[2]) for i in results if i[1] == first]))
print(second," ".join( [str(i[2]) for i in results if i[1] == second]))
# bar.render()