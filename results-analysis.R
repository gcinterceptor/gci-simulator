library(ggplot2)
library(plotly)

servers_number <- list(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 50, 100)
availability_rate <- list("0.5", "1.0", "2.0", "3.0", "4.0", "5.0", "6.0", "7.0", "8.0", "9.0", "10.0", "50.0")
communication_rate <- list("0.025")
load <- list("low", "high")

low_list <- list(list(), list(), list(), list())
count_low = 1
high_list <- list(list(), list(), list(), list())
count_high = 1

for(sn in servers_number) {
  for(ar in availability_rate) {
    for(cr in communication_rate) {
      for(ld in load) {
        baseline_scenario_file_path <- paste("results-1/1/request", sn, "baseline", ld, ar, paste(cr, ".csv", sep = ""), sep = "_")
        control_scenario_file_path <- paste("results-1/1/request", sn, "control", ld, ar, paste(cr, ".csv", sep = ""), sep = "_")
        
        if(file.exists(baseline_scenario_file_path) && file.exists(control_scenario_file_path)) {
          baseline <- read.csv(baseline_scenario_file_path, sep = ",", stringsAsFactors = FALSE)
          control <- read.csv(control_scenario_file_path, sep = ",", stringsAsFactors = FALSE)
          
          baseline_percentiles <- quantile(baseline$latency_time, c(.50, .99, .999))
          control_percentiles <- quantile(control$latency_time, c(.50, .99, .999))

          control_is_better <- TRUE
          for(i in 2:3) {
            if(control_percentiles[i] > baseline_percentiles[i]) {
              control_is_better <- FALSE
            }
          }
          
          color <- NA
          if(control_is_better) {
            if(control_percentiles[1] <= baseline_percentiles[1]) {
              color <- "50th, 99th, 99.9th"   # Median, 99th percentil and 99.9th percentile decreased (Median can be equal)
            } else {
              color <- "99th, 99.9th"         # 99th percentil and 99.9th percentile decreased
            }
          } else {
            color <- "None"                   # None decreased
          }
          
          if(ld == "low") {
            low_list[[1]][[count_low]] <- sn
            low_list[[2]][[count_low]] <- ar
            low_list[[3]][[count_low]] <- cr
            low_list[[4]][[count_low]] <- color
            count_low = count_low + 1
          } else {
            high_list[[1]][[count_high]] <- sn
            high_list[[2]][[count_high]] <- ar
            high_list[[3]][[count_high]] <- cr
            high_list[[4]][[count_high]] <- color
            count_high = count_high + 1
          }
          
          print(paste(sn, ar, cr, ld, color))
          print(baseline_percentiles)
          print(control_percentiles)
        }
      }
    }
  }
}

servers_number_v <- as.numeric(unlist(low_list[[1]], use.names = FALSE))
availability_rate_v <- as.numeric(unlist(low_list[[2]], use.names = FALSE))
communication_rate_v <- as.numeric(unlist(low_list[[3]], use.names = FALSE))
color_v <- unlist(low_list[[4]], use.names = FALSE)

low_results <- data.frame(servers_number_v, availability_rate_v, communication_rate_v, color_v, stringsAsFactors = FALSE)

servers_number_v <- as.numeric(unlist(high_list[[1]], use.names = FALSE))
availability_rate_v <- as.numeric(unlist(high_list[[2]], use.names = FALSE))
communication_rate_v <- as.numeric(unlist(high_list[[3]], use.names = FALSE))
color_v <- unlist(high_list[[4]], use.names = FALSE)

high_results <- data.frame(servers_number_v, availability_rate_v, communication_rate_v, color_v, stringsAsFactors = FALSE)

plot_ly(low_results, x = ~servers_number_v, y = ~availability_rate_v, 
        color = ~factor(color_v), 
        marker = list(size = 5)) %>% 
  layout(title = 'Low Load Results',
         scene = list(xaxis = list(title = 'SR'),
                      yaxis = list(title = 'AR')))

plot_ly(high_results, x = ~servers_number_v, y = ~availability_rate_v, 
        color = ~factor(color_v), 
        marker = list(size = 5)) %>% 
  layout(title = 'High Load Results',
         scene = list(xaxis = list(title = 'SR'),
                      yaxis = list(title = 'AR')))

