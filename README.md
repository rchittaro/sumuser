# Introduction (Work in Progress)

This is a simple program that will summarize all the data from a (specifically formatted) weekly log file as well as split out specific portions of the weekly 
log file into a user id specific output folder and set of files.

## Download and Install Go 

If you want to build and run this directly from the source code you will need to install golang

This program is written in Golang and will need to be installed and assumes go is in your path. Here are the instructions to install Golang. Make sure to the correct version for your
platform (Mac/Windows/Linux). https://go.dev/doc/install


## How to Run sumuser

- Open a terminal in the folder where you put the sumuser application
- It is expected there is a folder called 'data' within the application folder
- In the 'data' folder put your weekly log file and call it 'weekly_logs.csv'
- From the terminal simply run sumuser and give it the user_id you want to summarize. For example, for user_id 55 <br>
   % <b> go run . 55 </b>

## Output

The application will produce the following folder output for the example user_id '55'. 


output/ <br>
└── 55 <br>
    ├── calories (raw data for calories separated by week) <br>
    ├── flat_user_55.csv (Summarized flattened user id file. This file will over time contain more and more summarized data) <br>
    ├── heart (raw data for heart separated by week) <br>
    ├── sleep (raw data for sleep separated by week) <br>
    └── user_55.csv (all raw user_id 55 data) <br>
