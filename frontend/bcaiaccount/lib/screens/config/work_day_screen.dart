import 'package:smlaicloud/bloc/work_day/work_day_bloc.dart';
import 'package:smlaicloud/model/work_day_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class WorkDayScreen extends StatefulWidget {
  const WorkDayScreen({super.key});

  @override
  State<WorkDayScreen> createState() => WorkDayScreenState();
}

class WorkDayScreenState extends State<WorkDayScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController colorController = TextEditingController();
  Color currentColor = global.theme.negativeHighlightTextColor;
  String selectGuid = "";
  bool canNotSave = false;
  String time1 = "1";
  String time2 = "1";
  bool timeCheck = false;
  List<WorkDayModel> screenData = [
    WorkDayModel(
      code: 'mon',
      name: 'monday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'tue',
      name: 'tuesday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'wed',
      name: 'wendesday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'thu',
      name: 'thursday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'fri',
      name: 'friday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'sat',
      name: 'saturday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
    WorkDayModel(
      code: 'sun',
      name: 'sunday',
      isactive: false,
      fullday: false,
      worktimes: [
        WorkTimeModel(
          starttime: "08:00",
          endtime: "17:00",
          start: "800",
          end: "1700",
        ),
      ],
    ),
  ];
  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);
    loadDataList();
  }

  @override
  void initState() {
    setSystemLanguageList();

    super.initState();
  }

  void loadDataList() {
    context.read<WorkDayBloc>().add(const WorkDayLoad());
  }

  void changeColor(Color color) {
    setState(() => currentColor = color);
  }

  void addTimeRange(String code) {
    WorkDayModel wd = screenData.firstWhere((ele) => ele.code == code);
    if (wd.worktimes.length < 3) {
      setState(() {
        wd.worktimes.add(
          WorkTimeModel(
            starttime: "00:00",
            endtime: "00:00",
            start: "0000",
            end: "0000",
          ),
        );
      });
    }
  }

  void removeTimeRange(String code, int index) {
    WorkDayModel wd = screenData.firstWhere((ele) => ele.code == code);

    setState(() {
      wd.worktimes.removeAt(index);
    });
  }

  @override
  Widget build(BuildContext context) {
    int timefix1 = 1;
    int timefix2 = 1;

    List<Widget> formData = [];
    for (var workday in screenData) {
      List<Widget> listTime = [];
      workday.worktimes.asMap().forEach((index, time) {
        if (workday.isactive) {
          listTime.add(
            Container(
              margin: const EdgeInsets.all(2),
              alignment: Alignment.centerLeft,
              child: (!workday.fullday)
                  ? Column(
                      children: [
                        Row(
                          children: [
                            SizedBox(
                              width: 90,
                              height: 35,
                              child: TextField(
                                textAlign: TextAlign.center,
                                controller: TextEditingController(
                                  text: time.starttime,
                                ),
                                onTap: (() {
                                  //   DatePicker.showTimePicker(
                                  //     context,
                                  //     showSecondsColumn: false,
                                  //     showTitleActions: true,
                                  //     locale: LocaleType.th,
                                  //     onConfirm: (date) {
                                  //       String dateSplit = date.toString().split(' ')[1];
                                  //       List<String> timeSplit = dateSplit.split(':');

                                  //       setState(() {
                                  //         workday.worktimes[index].start = "${timeSplit[0]}${timeSplit[1]}";
                                  //         // if (int.parse(workday.worktimes[0].start) >=
                                  //         //     int.parse(workday.worktimes[1].start)) {
                                  //         //   // print("ตรงเงื่อนไข");
                                  //         // }
                                  //         // // print(workday.worktimes[0].starttime);
                                  //         // // print(workday.worktimes[1].starttime);
                                  //         // // print(workday.worktimes[2].starttime);
                                  //         time.starttime = "${timeSplit[0]}:${timeSplit[1]}";

                                  //         time.starttime = "${timeSplit[0]}:${timeSplit[1]}";

                                  //         time1 = "${timeSplit[0]}${timeSplit[1]}";
                                  //         // log(time.starttime + time.endtime);
                                  //         // timefix1 = int.parse(time1);
                                  //         // int.parse(time1);
                                  //         // if (int.parse(time1) > int.parse(time2)) {
                                  //         //   setState(() {
                                  //         //     Worktime.starttime + "Empty";
                                  //         //     Worktime.endtime =
                                  //         //         Worktime.endtime + "Empty";
                                  //         //     timecheck = true;
                                  //         //     log(timecheck.toString());
                                  //         //   });
                                  //         // } else if (int.parse(time1) <
                                  //         //     int.parse(time2)) {
                                  //         //   setState(() {
                                  //         //     timecheck = false;
                                  //         //   });

                                  //         //   log(timecheck.toString());
                                  //         // }
                                  //         if (workday.worktimes.length > 1 && workday.worktimes.length == 2) {
                                  //           if (int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].end) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("1");
                                  //             canNotSave = false;
                                  //           } else if (int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[0].start) < int.parse(workday.worktimes[1].end) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print(int.parse(workday.worktimes[1].end));
                                  //             // print("2");
                                  //             canNotSave = false;
                                  //             // print(int.parse(workday.worktimes[1].start));
                                  //           } else {
                                  //             // print("3XXX");
                                  //             setState(() {
                                  //               time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //               canNotSave = true;
                                  //             });
                                  //           }
                                  //         } else if (workday.worktimes.length > 1 && workday.worktimes.length == 3) {
                                  //           if (int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].end) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("31");
                                  //             canNotSave = false;
                                  //           }
                                  //           if (int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].end) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("31");
                                  //             canNotSave = false;
                                  //           } else if (int.parse(workday.worktimes[2].start) > int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[2].end) > int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("32");
                                  //             canNotSave = false;
                                  //           } else {
                                  //             // print("33");

                                  //             setState(() {
                                  //               time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //               canNotSave = true;
                                  //             });
                                  //           }
                                  //           if (int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].end) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("C1");
                                  //             canNotSave = false;
                                  //           } else if (int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[1].end) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("C2");
                                  //             canNotSave = false;
                                  //           } else {
                                  //             // print("3C");
                                  //             time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //             canNotSave = true;
                                  //           }
                                  //         }
                                  //         if (time.endtime.contains("!")) {
                                  //           canNotSave = true;
                                  //         } else {
                                  //           canNotSave = false;
                                  //         }
                                  //       });
                                  //     },
                                  //     currentTime: DateTime.now(),
                                  //   );
                                }),
                                readOnly: true,
                                decoration: InputDecoration(
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  border: OutlineInputBorder(),
                                  labelText: global.language("on"),
                                  labelStyle: TextStyle(fontSize: 14.0),
                                ),
                              ),
                            ),
                            const SizedBox(width: 5),
                            SizedBox(
                              width: 90,
                              height: 35,
                              child: TextField(
                                textAlign: TextAlign.center,
                                controller: TextEditingController(
                                  text: time.endtime,
                                ),
                                onTap: (() {
                                  //   DatePicker.showTimePicker(
                                  //     context,
                                  //     showSecondsColumn: false,
                                  //     showTitleActions: true,
                                  //     locale: LocaleType.th,
                                  //     onConfirm: (date) {
                                  //       String dateSplit = date.toString().split(' ')[1];
                                  //       List<String> timeSplit = dateSplit.split(':');
                                  //       setState(() {
                                  //         workday.worktimes[index].end = "${timeSplit[0]}${timeSplit[1]}";

                                  //         // print(workday.worktimes[index].starttime + "-" + workday.worktimes[index].endtime);

                                  //         // print(workday.worktimes[index].starttime + "-" + workday.worktimes[index].endtime);
                                  //         // if (workday.worktimes[0].starttime ==
                                  //         //     workday.worktimes[0].starttime) {
                                  //         //   // print('เท่ากัน');
                                  //         // }
                                  //         time.endtime = "${timeSplit[0]}:${timeSplit[1]}";
                                  //         time2 = "${timeSplit[0]}${timeSplit[1]}";
                                  //         log(time.starttime + time.endtime);
                                  //         int.parse(time2);
                                  //         log("time" + time1);
                                  //         log("time" + time2);
                                  //         // if (int.parse(time1) > int.parse(time2)) {
                                  //         //   setState(() {
                                  //         //     time.starttime = time.starttime;
                                  //         //     log(time.starttime);
                                  //         //     time.endtime = ("------");
                                  //         //     timecheck = true;
                                  //         //     log(timecheck.toString());
                                  //         //   });
                                  //         // } else if ((int.parse(time2)) -
                                  //         //         (int.parse(time1)) <
                                  //         //     100) {
                                  //         //   setState(() {
                                  //         //     time.endtime = ("-------");
                                  //         //     log("เวลาใกล้เคียงกันเกินไป");
                                  //         //   });

                                  //         //   log(timecheck.toString());
                                  //         // }
                                  //         if (int.parse(workday.worktimes[0].start) < int.parse(workday.worktimes[0].end)) {
                                  //         } else {
                                  //           // print("ff");
                                  //           time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //         }
                                  //         if (workday.worktimes.length > 1 && workday.worktimes.length <= 2) {
                                  //           if (int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].end) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("1");
                                  //             canNotSave = false;
                                  //           } else if (int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[0].start) < int.parse(workday.worktimes[1].end) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("2");
                                  //             canNotSave = false;
                                  //           } else {
                                  //             // print("3");
                                  //             setState(() {
                                  //               time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //               canNotSave = true;
                                  //             });
                                  //           }
                                  //         } else if (workday.worktimes.length > 1 && workday.worktimes.length == 3) {
                                  //           if (int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].end) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("31");
                                  //             canNotSave = false;
                                  //           }
                                  //           if (int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].end) < int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("31");
                                  //             canNotSave = false;
                                  //             // print(canNotSave);
                                  //           } else if (int.parse(workday.worktimes[2].start) > int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[2].end) > int.parse(workday.worktimes[1].start) &&
                                  //               int.parse(workday.worktimes[2].start) < int.parse(workday.worktimes[2].end)) {
                                  //             // print("32");
                                  //             canNotSave = false;
                                  //           } else {
                                  //             // print("33");

                                  //             setState(() {
                                  //               time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //               canNotSave = true;
                                  //             });
                                  //           }
                                  //           if (int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].end) < int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("C1");
                                  //             canNotSave = false;
                                  //           } else if (int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) > int.parse(workday.worktimes[0].end) &&
                                  //               int.parse(workday.worktimes[1].end) > int.parse(workday.worktimes[0].start) &&
                                  //               int.parse(workday.worktimes[1].start) < int.parse(workday.worktimes[1].end)) {
                                  //             // print("C2");
                                  //             canNotSave = false;
                                  //           } else {
                                  //             // print("3C");
                                  //             time.endtime = "${timeSplit[0]}:${timeSplit[1]}!";
                                  //             canNotSave = true;
                                  //           }
                                  //         }
                                  //         if (time.endtime.contains("!")) {
                                  //           canNotSave = true;
                                  //         } else {
                                  //           canNotSave = false;
                                  //         }
                                  //         // if (int.parse(time.starttime) ==
                                  //         //     time.endtime) {}
                                  //         // timefix1 = int.parse(time2);
                                  //       });
                                  //     },
                                  //     currentTime: DateTime.now(),
                                  //   );
                                }),
                                readOnly: true,
                                decoration: InputDecoration(
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  border: OutlineInputBorder(),
                                  labelText: global.language("off"),
                                  labelStyle:
                                      (time.starttime == time.endtime ||
                                          time.endtime == ("------"))
                                      ? TextStyle(
                                          color: global
                                              .theme
                                              .inputTextBoxForceColor,
                                        )
                                      : TextStyle(fontSize: 14.0),
                                ),
                              ),
                            ),
                            Expanded(
                              child: (index != 0)
                                  ? Container(
                                      margin: const EdgeInsets.only(right: 5),
                                      alignment: Alignment.centerRight,
                                      child: IconButton(
                                        onPressed: () {
                                          removeTimeRange(workday.code, index);
                                        },
                                        icon: Icon(Icons.delete_outline),
                                      ),
                                    )
                                  : Container(),
                            ),
                          ],
                        ),
                        Container(
                          alignment: Alignment.centerLeft,
                          child:
                              (time.starttime == time.endtime &&
                                  time.endtime != "------")
                              ? Text(
                                  style: TextStyle(color: global.theme.negativeHighlightTextColor),
                                  " ${global.language("invalid_format")} ",
                                )
                              : Container(),
                        ),
                        Container(
                          alignment: Alignment.centerLeft,
                          child: (time.endtime == "------")
                              ? Text(
                                  global.language("invalid_format"),
                                  style: TextStyle(
                                    color: global.theme.negativeHighlightTextColor,
                                  ),
                                )
                              : Container(),
                        ),
                        Container(
                          alignment: Alignment.centerLeft,
                          child: (time.endtime.contains("!"))
                              ? Text(
                                  global.language("invalid_format"),
                                  style: TextStyle(
                                    color: global.theme.negativeHighlightTextColor,
                                  ),
                                )
                              : Container(),
                        ),
                        Container(
                          alignment: Alignment.centerLeft,
                          child: (time.endtime == "-------")
                              ? Text(
                                  global.language("invalid_format"),
                                  style: TextStyle(
                                    color: global.theme.negativeHighlightTextColor,
                                  ),
                                )
                              : Container(),
                        ),
                      ],
                    )
                  : null,
            ),
          );
        }
      });
      formData.add(
        Card(
          elevation: 10,
          child: Padding(
            padding: EdgeInsets.all(8.0),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      global.language(workday.name),
                      style: TextStyle(fontSize: 18),
                    ),
                    Switch(
                      value: workday.isactive,
                      onChanged: (value) {
                        setState(() {
                          workday.isactive = value;
                        });
                      },
                    ),
                  ],
                ),
                Divider(height: 1),
                Container(
                  margin: EdgeInsets.only(top: 5),
                  alignment: Alignment.centerLeft,
                  child: (!workday.isactive)
                      ? Text(
                          global.language("off"),
                          style: TextStyle(
                            fontSize: 17,
                            color: global.theme.negativeHighlightTextColor,
                          ),
                        )
                      : null,
                ),
                Container(
                  alignment: Alignment.centerLeft,
                  child: (workday.isactive)
                      ? Row(
                          children: <Widget>[
                            Checkbox(
                              value: workday.fullday,
                              onChanged: (bool? newValue) {
                                setState(() {
                                  workday.fullday = newValue!;
                                });
                              },
                            ),
                            Text("24 ${global.language("hours")}"),
                          ],
                        )
                      : null,
                ),
                Column(children: listTime),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  alignment: Alignment.centerLeft,
                  child:
                      (!workday.fullday &&
                          workday.isactive &&
                          workday.worktimes.length < 3)
                      ? TextButton(
                          onPressed: () {
                            addTimeRange(workday.code);
                          },
                          child: Text(
                            global.language("duration"),
                            style: TextStyle(color: global.theme.infoHighlightTextColor),
                          ),
                        )
                      : null,
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topRight,
          end: Alignment.bottomLeft,
          colors: [global.theme.infoHighlightColor, global.theme.columnHeaderColor],
        ),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          return Scaffold(
            backgroundColor: Colors.transparent,
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              automaticallyImplyLeading: false,
              title: Text(global.language('workday')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
              actions: <Widget>[
                Padding(
                  padding: EdgeInsets.only(right: 10.0),
                  child: IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    onPressed: () {
                      if (canNotSave == false) {
                        saveOrUpdateData();
                      } else {
                        showDialog(
                          context: context,
                          builder: (BuildContext context) {
                            return AlertDialog(
                              title: Text(global.language('cannot_save')),
                              content: Text(
                                global.language('please_check_time_format'),
                              ),
                              actions: <Widget>[
                                TextButton(
                                  child: Text(global.language('ok')),
                                  onPressed: () {
                                    Navigator.of(context).pop();
                                  },
                                ),
                              ],
                            );
                          },
                        );
                      }
                    },
                    icon: Icon(Icons.save, size: 26.0),
                  ),
                ),
              ],
            ),
            body: SingleChildScrollView(
              child: BlocListener<WorkDayBloc, WorkDayState>(
                listener: (context, state) {
                  if (state is WorkDayLoadSuccess) {
                    global.showInfoSnackBar(context, global.language('loadsuc'));
                    setState(() {
                      selectGuid = state.guidfixed;
                      screenData = state.workDay;
                    });
                  }
                  if (state is WorkDayLoadFailed) {
                    global.showInfoSnackBar(context, global.language('eror${state.message}"'));
                    // setState(() {
                    //   selectGuid = state.guidfixed;
                    //   screenData = state.workday;
                    // });
                  }
                  if (state is WorkDaySaveSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("save_success"),
                      global.theme.infoHighlightTextColor,
                    );
                    loadDataList();
                  }
                  if (state is WorkDaySaveFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("not_success_save ${state.message}"),
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                  if (state is WorkDayUpdateSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("updatesuc"),
                      global.theme.infoHighlightTextColor,
                    );
                    loadDataList();
                  }
                  if (state is WorkDayUpdateFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("not_edit_success ${state.message}"),
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                },
                child: Padding(
                  padding: const EdgeInsets.all(10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      SizedBox(
                        width: (constraints.maxWidth > 800) ? 450 : 320,
                        child: Column(children: formData),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  void saveOrUpdateData() {
    if (selectGuid.trim().isEmpty) {
      WorkDayListModel wdList = WorkDayListModel(workdays: screenData);
      context.read<WorkDayBloc>().add(WorkDaySave(workDays: wdList));
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    WorkDayListModel wdList = WorkDayListModel(workdays: screenData);
    context.read<WorkDayBloc>().add(
      WorkDayUpdate(guid: guid, workDays: wdList),
    );
  }
}
