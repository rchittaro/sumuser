# Introduction (Work in Progress)

This is a simple program that will summarize all the data from a (specifically formatted) weekly log file as well as split out specific portions of the weekly 
log file into a user id specific output folder and set of files.

## Download and Install Go - (Not Suggested)

If you want to build and run this directly from the source code you will need to install golang

This program is written in Golang and will need to be installed and assumes go is in your path. Here are the instructions to install Golang. Make sure to the correct version for your
platform (Mac/Windows/Linux). https://go.dev/doc/install

## Download sumuser Application Only - Suggested

There are a set of already compiled programs organized by the target host (Mac/Windows/Linux) where it will executed. Just grab the one that matches your 
laptop or host. 

bin <br>
├── Apple
│   ├── amd64
│   │   └── sumuser
│   └── arm64
│       └── sumuser (Likely for anyone with a recent Mac)
├── Linux
│   └── amd64
│       └── sumuser (Any 64-bit Linux)
└── Windows
    ├── amd64
    │   └── sumuser (Likely for anyone with a recent Windows laptop)
    └── x86
        └── sumuser


## How to Run sumuser

- Open a terminal in the folder where you put the sumuser application
- It is expected there is a folder called 'data' within the application folder
- In the 'data' folder put your weekly log file and call it 'weekly_logs.csv'
- From the terminal simply run sumuser and give it the user_id you want to summarize. For example, for user_id 55
   ./sumuser 55

## Output

The application will produce the following folder output for the example user_id '55'. 


output/
└── 55
    ├── calories (raw data for calories separated by week)
    │   ├── calories_user_55_week_17.txt
    │   ├── calories_user_55_week_23.txt
    │   ├── calories_user_55_week_24.txt
    │   ├── calories_user_55_week_28.txt
    │   ├── calories_user_55_week_31.txt
    │   └── calories_user_55_week_46.txt
    ├── flat_user_55.csv (Summarized flattened user id file. This file will over time contain more and more summarized data)
    ├── heart (raw data for heart separated by week)
    │   ├── heart_user_55_week_17.txt
    │   ├── heart_user_55_week_23.txt
    │   ├── heart_user_55_week_24.txt
    │   ├── heart_user_55_week_28.txt
    │   ├── heart_user_55_week_31.txt
    │   └── heart_user_55_week_46.txt
    ├── sleep (raw data for sleep separated by week)
    │   ├── sleep_user_55_week_16.txt
    │   ├── sleep_user_55_week_17.txt
    │   ├── sleep_user_55_week_18.txt
    │   ├── sleep_user_55_week_19.txt
    │   ├── sleep_user_55_week_21.txt
    │   ├── sleep_user_55_week_22.txt
    │   ├── sleep_user_55_week_23.txt
    │   ├── sleep_user_55_week_24.txt
    │   ├── sleep_user_55_week_25.txt
    │   ├── sleep_user_55_week_26.txt
    │   ├── sleep_user_55_week_28.txt
    │   ├── sleep_user_55_week_31.txt
    │   ├── sleep_user_55_week_33.txt
    │   ├── sleep_user_55_week_35.txt
    │   └── sleep_user_55_week_46.txt
    └── user_55.csv (all raw user_id 55 data)
