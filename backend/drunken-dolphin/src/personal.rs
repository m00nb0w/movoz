use crate::config::AppConfig;
use crate::error::AppError;
use chrono::{Local, NaiveDate};
use colored::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;

#[derive(Debug, Serialize, Deserialize)]
pub struct ExerciseRecord {
    pub count: u32,
    pub timestamp: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DailyRecord {
    pub pushups: Option<ExerciseRecord>,
    pub situps: Option<ExerciseRecord>,
    pub pullups: Option<ExerciseRecord>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PersonalData {
    pub fitness: HashMap<String, DailyRecord>,
    // Future: pub finance: FinanceData,
    // Future: pub chores: ChoreData,
}

pub struct Tracker {
    config: AppConfig,
    data: PersonalData,
}

impl Tracker {
    pub fn new(config: AppConfig) -> Result<Self, AppError> {
        let data = if config.data_file.exists() {
            let content = fs::read_to_string(&config.data_file)?;
            serde_json::from_str(&content).unwrap_or_else(|_| PersonalData {
                fitness: HashMap::new(),
            })
        } else {
            PersonalData {
                fitness: HashMap::new(),
            }
        };

        Ok(Tracker { config, data })
    }

    pub fn save(&self) -> Result<(), AppError> {
        let content = serde_json::to_string_pretty(&self.data)?;
        fs::write(&self.config.data_file, content)?;
        Ok(())
    }

    pub fn record_pushups(&mut self, count: u32, date: &str) -> Result<(), AppError> {
        let date_key = self.parse_date(date)?;
        let record = ExerciseRecord {
            count,
            timestamp: Local::now().format("%Y-%m-%d %H:%M:%S").to_string(),
        };

        let daily_record = self.data.fitness.entry(date_key.clone()).or_insert_with(|| DailyRecord {
            pushups: None,
            situps: None,
            pullups: None,
        });

        daily_record.pushups = Some(record);
        self.save()?;

        println!("{} {} push-ups recorded for {}! 💪", 
            "✓".green().bold(), 
            count.to_string().yellow().bold(), 
            date_key.blue().bold()
        );

        Ok(())
    }

    pub fn record_situps(&mut self, count: u32, date: &str) -> Result<(), AppError> {
        let date_key = self.parse_date(date)?;
        let record = ExerciseRecord {
            count,
            timestamp: Local::now().format("%Y-%m-%d %H:%M:%S").to_string(),
        };

        let daily_record = self.data.fitness.entry(date_key.clone()).or_insert_with(|| DailyRecord {
            pushups: None,
            situps: None,
            pullups: None,
        });

        daily_record.situps = Some(record);
        self.save()?;

        println!("{} {} sit-ups recorded for {}! 🏃", 
            "✓".green().bold(), 
            count.to_string().yellow().bold(), 
            date_key.blue().bold()
        );

        Ok(())
    }

    pub fn record_pullups(&mut self, count: u32, date: &str) -> Result<(), AppError> {
        let date_key = self.parse_date(date)?;
        let record = ExerciseRecord {
            count,
            timestamp: Local::now().format("%Y-%m-%d %H:%M:%S").to_string(),
        };

        let daily_record = self.data.fitness.entry(date_key.clone()).or_insert_with(|| DailyRecord {
            pushups: None,
            situps: None,
            pullups: None,
        });

        daily_record.pullups = Some(record);
        self.save()?;

        println!("{} {} pull-ups recorded for {}! 🏋️", 
            "✓".green().bold(), 
            count.to_string().yellow().bold(), 
            date_key.blue().bold()
        );

        Ok(())
    }

    fn parse_date(&self, date: &str) -> Result<String, AppError> {
        match date.to_lowercase().as_str() {
            "today" => {
                let today = Local::now().date_naive();
                Ok(today.format("%Y-%m-%d").to_string())
            }
            "yesterday" => {
                let yesterday = Local::now().date_naive() - chrono::Duration::days(1);
                Ok(yesterday.format("%Y-%m-%d").to_string())
            }
            _ => {
                // Try to parse as YYYY-MM-DD
                if let Ok(parsed_date) = NaiveDate::parse_from_str(date, "%Y-%m-%d") {
                    Ok(parsed_date.format("%Y-%m-%d").to_string())
                } else {
                    Err(AppError::InvalidDate(format!(
                        "Invalid date format: {}. Use 'today', 'yesterday', or YYYY-MM-DD",
                        date
                    )))
                }
            }
        }
    }
}
