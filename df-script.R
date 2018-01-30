library(ggplot2)
library(dplyr)

rp <- 1

df <- read.csv("/home/pfelipe/gci-simulator/results-output-1.csv", stringsAsFactors = FALSE)

colnames(df) <- c("SN", "AR", "Load", "Statistic", "Value")

df_low <- df[which(df$Load == 'low' & df$SN != 100), ]
df_high <- df[which(df$Load == 'high' & df$SN != 100), ]

plot <- ggplot(df_low, aes(x = factor(Statistic, levels = 
                                        c("Mean", "50%", "75%", "90%", "99%", "99.9%", "99.99%", "99.999%", "100%")), 
                           y = Value)) + 
  geom_point() + xlab("Statistic Type") + ylab("Improvement Percentage") + 
  theme(axis.text.x = element_text(angle = 90, hjust = 1)) + 
  facet_grid(SN ~ AR, labeller=label_both)

ggsave(paste("low-load-", rp, ".png", sep=""), plot, width = 14, height = 8)

plot <- ggplot(df_high, aes(x = factor(Statistic, levels = 
                                         c("Mean", "50%", "75%", "90%", "99%", "99.9%", "99.99%", "99.999%", "100%")), 
                            y = Value)) + 
  geom_point() + xlab("Statistic Type") + ylab("Improvement Percentage") + 
  theme(axis.text.x = element_text(angle = 90, hjust = 1)) + 
  facet_grid(SN ~ AR, labeller=label_both)

ggsave(paste("high-load-", rp, ".png", sep=""), plot, width = 14, height = 8)

