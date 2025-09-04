use thiserror::Error;
use std::io;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("Configuration error: {0}")]
    Config(String),
    
    #[error("JSON parsing error: {0}")]
    Json(#[from] serde_json::Error),
    
    #[error("IO error: {0}")]
    Io(#[from] io::Error),
    
    #[error("Invalid date format: {0}")]
    InvalidDate(String),
    
    #[error("Invalid exercise type: {0}")]
    InvalidExercise(String),
}

impl From<String> for AppError {
    fn from(err: String) -> Self {
        AppError::Config(err)
    }
}
