pub mod pushups;
pub mod situps;
pub mod pullups;

// Re-export specific functions
pub use pushups::execute as pushups_execute;
pub use situps::execute as situps_execute;
pub use pullups::execute as pullups_execute;
