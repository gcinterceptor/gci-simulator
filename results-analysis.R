library(ggplot2)
library(plotly)

servers_number <- list(2, 4, 8, 50)
availability_rate <- list("0.1", "0.5", "1.0", "2.0", "10.0")
communication_rate <- list("0.025")
load <- list("low", "high")
statistic <- c(.50, .75, .90, .99, .999, .9999, .99999, .100)

statistic_names <- c("50%", "75%", "90%", "99%", "99.9%", "99.99%", "99.999%", "100%", "Mean")

for(rp in 1:10) {
  result_list <- list(list(), list(), list(), list(), list(), list())
  count <- 1
  for(sn in servers_number) {
    for(ar in availability_rate) {
      for(cr in communication_rate) {
        for(ld in load) {
          control_scenario_file_path <- paste(paste("results-only-gci", rp, "request", sep="/"), sn, "control", ld, ar, paste(cr, ".csv", sep = ""), sep = "_")
          
          if(file.exists(control_scenario_file_path)) {
            control <- read.csv(control_scenario_file_path, sep = ",", stringsAsFactors = FALSE)
            
            for(st_id in 1:length(statistic)) {
              result_list[[1]][[count]] <- sn
              result_list[[2]][[count]] <- ar
              result_list[[3]][[count]] <- ld
              result_list[[4]][[count]] <- statistic_names[st_id]
              
              control_percentil <- quantile(control$latency_time, statistic[st_id])
              
              result_list[[5]][[count]] <- control_percentil
              count <- count + 1
              
              print(paste(rp, sn, ar, cr, ld, statistic_names[st_id], control_percentil))
            }
            
            result_list[[1]][[count]] <- sn
            result_list[[2]][[count]] <- ar
            result_list[[3]][[count]] <- ld
            result_list[[4]][[count]] <- "Mean"
            
            control_mean <- mean(control$latency_time)
            
            result_list[[5]][[count]] <- control_mean
            count <- count + 1
            
            print(paste(rp, sn, ar, cr, ld, "Mean", control_mean))
          }
        }
      }
    }
  }
  
  gc()
  
  servers_number_v <- as.numeric(unlist(result_list[[1]], use.names = FALSE))
  availability_rate_v <- as.numeric(unlist(result_list[[2]], use.names = FALSE))
  load_v <- unlist(result_list[[3]], use.names = FALSE)
  statistic_name_v <- unlist(result_list[[4]], use.names = FALSE)
  statistic_improve_v <- as.numeric(unlist(result_list[[5]], use.names = FALSE))
  
  df <- data.frame(Servers_number = servers_number_v, Availability_rate = availability_rate_v, Load = load_v, 
                   Statistic = statistic_name_v, Value = statistic_improve_v, stringsAsFactors = FALSE)
  
  write.csv(df, paste("results-output-control-", rp, ".csv", sep = ""), row.names = FALSE)
  
}