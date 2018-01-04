addCommunicationTime <- function(data, redirect_time) {
  newData <- data
  newData$latency_time <- data$latency_time + ((data$redirects - 1) * redirect_time)
  return(newData)
}

args <- commandArgs(trailingOnly=TRUE)

baselineScenarioFilePath <- args[1]
controlScenarioFilePath <- args[2]

baseline <- read.csv(baselineScenarioFilePath, sep = ",", stringsAsFactors = FALSE)
control <- read.csv(controlScenarioFilePath, sep = ",", stringsAsFactors = FALSE)

unavailable_avg_time <- 1
maxRedirectPercent <- -1
for(redirect_percent in seq(0.05, 1, 0.05)) {
  redirect_time <- redirect_percent * unavailable_avg_time
  
  baselineData <- addCommunicationTime(baseline, redirect_time)
  controlData <- addCommunicationTime(control, redirect_time)
  
  baselinePercentiles <- quantile(baselineData$latency_time, c(.50, .99, .999))
  controlPercentiles <- quantile(controlData$latency_time, c(.50, .99, .999))
  
  print(paste("Redirect Percent :: ", redirect_percent))
  print(baselinePercentiles)
  print(controlPercentiles)
  
  controlIsBetter <- TRUE
  for(i in 2:3) {
    if(controlPercentiles[i] > baselinePercentiles[i]) {
      controlIsBetter <- FALSE
    }
  }
  if(controlIsBetter) {
    maxRedirectPercent <- redirect_percent
  }
}

print(paste("Max Redirect Percent = ", maxRedirectPercent))