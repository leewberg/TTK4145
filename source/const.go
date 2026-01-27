package elevio

const NUM_ELEVATORS = 3
const NUM_FLOORS = 4

var MY_ID int
var LocalElevator Elevator

const INF = 2147483647        // 32 bit signed integer limit
const ELEVATOR_TIMEOUT = 6000 // ms
const TRAVEL_TIME = 2500      // ms
const DOOR_OPEN_TIME = 3000   // ms
const BIDDING_TIME = 300      // ms
const BIDDING_MIN_RAISE = 100 // ms
const IS_ALONE_TIMEOUT = 500  // ms
const BROADCAST_PERIOD = 50   // ms
