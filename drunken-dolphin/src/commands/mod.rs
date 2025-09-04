pub mod pushups;
pub mod situps;

// Re-export specific functions
pub use pushups::execute as pushups_execute;
pub use situps::execute as situps_execute;
