// Package frapi provides the entry points for the graphql fr api.
// The tests are minimal and simply check the functionality of graphql calls.
// More extensive tests are performed by higher level client javascript calls.
/*
Example calls:

new property:

mutation NewProperty
{
  createProperty1(input: {name: "newThree", currency: USD, memberrate: 40.00, allownonmembers: false, nonmemberrate: 80.00, timeseconds: 0}) {
    id
    name
  }
}

query properties:

{
  properties1 {
    id
    name
    currency
    memberrate
    nonmemberrate
    allownonmembers
  }
}

new reservation:

mutation NewReservation
{
  createReservation(id: "01a79cdf-668a-4d33-813c-46a9e24e7c60", input: {madefor: "bill", madeby: "sally", startdate: "now", enddate: "later", member: false, rates: [50.0]}) {
    id
  }
}

*/
package frapi
