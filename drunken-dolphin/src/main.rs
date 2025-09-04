use clap::{Parser, Subcommand};
use colored::*;
use std::process;

mod commands;
mod config;
mod error;
mod personal;

use commands::{pushups_execute, situps_execute};
use config::AppConfig;
use error::AppError;

#[derive(Parser)]
#[clap(name = "drunken-dolphin")]
#[clap(about = "🐬 A personal CLI assistant for life management")]
#[clap(version)]
struct Cli {
    #[clap(subcommand)]
    command: Commands,
    
    /// Data file path for storing personal data
    #[clap(long, default_value = "personal_data.json")]
    data_file: String,
}

#[derive(Subcommand)]
enum Commands {
    /// Fitness tracking commands
    Fitness {
        #[clap(subcommand)]
        command: FitnessCommands,
    },
    
    /// Chore management commands
    Chore {
        #[clap(subcommand)]
        command: ChoreCommands,
    },
}

#[derive(Subcommand)]
enum FitnessCommands {
    /// Record push-up count
    Pushups {
        /// Number of push-ups completed
        count: u32,
        
        /// Date to record for: today, yesterday, or specific date (YYYY-MM-DD)
        #[clap(short, long, default_value = "today")]
        date: String,
    },
    
    /// Record sit-up count
    Situps {
        /// Number of sit-ups completed
        count: u32,
        
        /// Date to record for: today, yesterday, or specific date (YYYY-MM-DD)
        #[clap(short, long, default_value = "today")]
        date: String,
    },
    
    /// Quick shortcut: Record both exercises for today
    Today {
        /// Number of push-ups completed
        pushups: u32,
        
        /// Number of sit-ups completed
        situps: u32,
    },
    
    /// Quick shortcut: Record both exercises for yesterday
    Yesterday {
        /// Number of push-ups completed
        pushups: u32,
        
        /// Number of sit-ups completed
        situps: u32,
    },
}

#[derive(Subcommand)]
enum ChoreCommands {
    /// Add a new chore
    Add {
        /// Description of the chore
        description: String,
        
        /// Priority level (low, medium, high)
        #[clap(short, long, default_value = "medium")]
        priority: String,
        
        /// Due date for the chore
        #[clap(short, long)]
        due_date: Option<String>,
    },
    
    /// Mark a chore as complete
    Complete {
        /// Description of the chore to complete
        description: String,
    },
}

fn main() {
    if let Err(err) = run() {
        eprintln!("{} {}", "Error:".red().bold(), err);
        process::exit(1);
    }
}

fn run() -> Result<(), AppError> {
    let cli = Cli::parse();
    
    // Initialize configuration
    let config = AppConfig::new(cli.data_file)?;
    
    // Initialize personal data tracker
    let mut tracker = personal::Tracker::new(config)?;
    
    match cli.command {
        Commands::Fitness { command } => {
            match command {
                FitnessCommands::Pushups { count, date } => {
                    pushups_execute(&mut tracker, count, &date)?;
                }
                FitnessCommands::Situps { count, date } => {
                    situps_execute(&mut tracker, count, &date)?;
                }
                FitnessCommands::Today { pushups, situps } => {
                    pushups_execute(&mut tracker, pushups, "today")?;
                    situps_execute(&mut tracker, situps, "today")?;
                }
                FitnessCommands::Yesterday { pushups, situps } => {
                    pushups_execute(&mut tracker, pushups, "yesterday")?;
                    situps_execute(&mut tracker, situps, "yesterday")?;
                }
            }
        }
        
        Commands::Chore { command } => {
            match command {
                ChoreCommands::Add { description, priority, due_date } => {
                    println!("{} 🧹 Chore management coming soon!", "⚠️".yellow().bold());
                    println!("Description: {}", description);
                    println!("Priority: {}", priority);
                    if let Some(due) = due_date {
                        println!("Due date: {}", due);
                    }
                }
                ChoreCommands::Complete { description } => {
                    println!("{} ✅ Chore completion coming soon!", "⚠️".yellow().bold());
                    println!("Completing: {}", description);
                }
            }
        }
    }
    
    Ok(())
}
