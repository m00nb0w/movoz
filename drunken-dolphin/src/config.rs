use crate::error::AppError;
use std::path::PathBuf;

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub data_file: PathBuf,
}

impl AppConfig {
    pub fn new(data_file: String) -> Result<Self, AppError> {
        let data_file = PathBuf::from(data_file);
        
        // Create parent directory if it doesn't exist
        if let Some(parent) = data_file.parent() {
            if !parent.exists() {
                std::fs::create_dir_all(parent)
                    .map_err(|e| AppError::Config(format!("Failed to create directory: {}", e)))?;
            }
        }
        
        Ok(AppConfig { data_file })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::tempdir;

    #[test]
    fn test_config_new_with_valid_path() {
        let temp_dir = tempdir().unwrap();
        let data_file = temp_dir.path().join("fitness.json");
        let config = AppConfig::new(data_file.to_string_lossy().to_string());
        assert!(config.is_ok());
        
        let config = config.unwrap();
        assert_eq!(config.data_file, data_file);
    }
}
