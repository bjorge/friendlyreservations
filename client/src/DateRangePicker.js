import React from 'react';
import Helmet from 'react-helmet';
import DayPicker, { DateUtils } from 'react-day-picker';
import { inject, observer } from "mobx-react";
import 'react-day-picker/lib/style.css';


class DateRangePicker extends React.Component {
  static defaultProps = {
    numberOfMonths: 1,
  };
  constructor(props) {
    super(props);
    this.handleDayClick = this.handleDayClick.bind(this);
    this.handleResetClick = this.handleResetClick.bind(this);
    this.state = this.getInitialState();
  }
  getInitialState() {
    return {
      from: undefined,
      to: undefined
    };
  }
  handleDayClick(day, modifiers) {
    if (modifiers.disabled !== true) {
      const range = DateUtils.addDayToRange(day, this.state);
      if (range.from === range.to) {
        this.handleResetClick()
      } else {
        this.setState(range);
        this.props.onUpdateRange(range);
      }
    }
  }

  handleResetClick() {
    this.setState(this.getInitialState());
    this.props.onUpdateRange({ to: undefined, from: undefined });
  }
  render() {
    const { from, to } = this.state;
    const modifiers = { start: from, end: to };

    // console.log("checkinDisabled:")
    // console.log(this.props.checkinDisabled)
    // console.log("checkoutDisabled:")
    // console.log(this.props.checkoutDisabled)

    // convert iso date strings from service to javascript dates
    var disabledInRanges = JSON.parse(JSON.stringify(this.props.checkinDisabled));
    var disabledOutRanges = JSON.parse(JSON.stringify(this.props.checkoutDisabled));
    var arrayInLength = disabledInRanges.length;
    var key, date, i
    for (i = 0; i < arrayInLength; i++) {
      for (key in disabledInRanges[i]) {
        if (disabledInRanges[i][key]) {
          date = new Date(disabledInRanges[i][key].replace(/-/g, '/'))
          disabledInRanges[i][key] = date
        }
      }
    }
    var arrayOutLength = disabledOutRanges.length;
    for (i = 0; i < arrayOutLength; i++) {
      for (key in disabledOutRanges[i]) {
        if (disabledOutRanges[i][key]) {
          date = new Date(disabledOutRanges[i][key].replace(/-/g, '/'))
          disabledOutRanges[i][key] = date
        }
      }
    }

    // by default disable invalid checkin dates
    var disabledRanges = disabledInRanges
    if (from) {
      // ok we've selected the checkin date, so now disable invalid checkout dates

      // clear out the disabled ranges
      disabledRanges = []

      // remove days before 'from' now that 'from' has been selected
      disabledRanges.push({ before: from })

      // add disabled out days after the from date
      for (i = 0; i < arrayOutLength; i++) {
        for (key in disabledOutRanges[i]) {
          var checkoutDate = disabledOutRanges[i][key]
          if (key === "from" && from <= checkoutDate) {
            disabledRanges.push({ from: checkoutDate, to: disabledOutRanges[i]['to'] })
            disabledRanges.push({ after: disabledOutRanges[i]['to'] })
          }          
          if (key === "after" && from <= checkoutDate) {
            disabledRanges.push({ after: checkoutDate })
          }
        }
      }
    }

    return (
      <div className="RangeExample" >
        < p >
          {' '}
          {from && (
            <button className="link" onClick={this.handleResetClick}>
              Reset Dates
                    </button>
          )}
        </p >
        <DayPicker
          className="Selectable"
          onDayClick={this.handleDayClick}
          numberOfMonths={this.props.numberOfMonths}
          selectedDays={[from, { from, to }]}
          modifiers={modifiers}
          disabledDays={disabledRanges}
        />
        <Helmet>
          <style>{`
    .Selectable .DayPicker-Day--selected:not(.DayPicker-Day--start):not(.DayPicker-Day--end):not(.DayPicker-Day--outside) {
      background-color: #f0f8ff !important;
      color: #4a90e2;
    }
    .Selectable .DayPicker-Day {
      border-radius: 0 !important;
    }
    .Selectable .DayPicker-Day--start {
      border-top-left-radius: 50% !important;
      border-bottom-left-radius: 50% !important;
    }
    .Selectable .DayPicker-Day--end {
      border-top-right-radius: 50% !important;
      border-bottom-right-radius: 50% !important;
    }
  `}</style>
        </Helmet>
      </div >
    );
  }
}







export default inject('appStateStore')(observer(DateRangePicker))
