/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

pub mod gen_go;

use anyhow::{bail, Context};
use camino::{Utf8Path, Utf8PathBuf};
use cargo_metadata::{MetadataCommand, Package};
use clap::Parser;
use fs_err::{self as fs};
use gen_go::{generate_go_bindings, Config};
use std::{
    collections::{HashMap, HashSet},
    process::Command,
};
use uniffi_bindgen::{
    interface::ComponentInterface, macro_metadata, BindingGenerator, BindingsConfig,
};

#[derive(Parser)]
#[clap(name = "uniffi-bindgen")]
#[clap(version = clap::crate_version!())]
#[clap(propagate_version = true)]
struct Cli {
    /// Directory in which to write generated files. Default is same folder as .udl file.
    #[clap(long, short)]
    out_dir: Option<Utf8PathBuf>,

    /// Do not try to format the generated bindings.
    #[clap(long, short)]
    no_format: bool,

    /// Path to the optional uniffi config file. If not provided, uniffi-bindgen will try to guess it from the UDL's file location.
    #[clap(long, short)]
    config: Option<Utf8PathBuf>,

    /// Extract proc-macro metadata from a native lib (cdylib or staticlib) for this crate.
    #[clap(long, short)]
    lib_file: Option<Utf8PathBuf>,

    /// Pass in a cdylib path rather than a UDL file
    #[clap(long = "library")]
    library_mode: bool,

    /// When `--library` is passed, only generate bindings for one crate.
    /// When `--library` is not passed, use this as the crate name instead of attempting to
    /// locate and parse Cargo.toml.
    #[clap(long = "crate")]
    crate_name: Option<String>,

    /// Path to the UDL file, or cdylib if `library-mode` is specified
    source: Utf8PathBuf,
}

struct BindingGeneratorGo {
    try_format_code: bool,
}

impl uniffi_bindgen::BindingGenerator for BindingGeneratorGo {
    type Config = gen_go::Config;

    fn write_bindings(
        &self,
        ci: ComponentInterface,
        config: Self::Config,
        out_dir: &Utf8Path,
    ) -> anyhow::Result<()> {
        let bindings_path = full_bindings_path(&config, &ci, out_dir);
        fs::create_dir_all(&bindings_path)?;
        let go_file = bindings_path.join(format!("{}.go", ci.namespace()));
        let (header, c_file_content, wrapper) = generate_go_bindings(&config, &ci)?;
        fs::write(&go_file, wrapper)?;

        let header_file = bindings_path.join(config.header_filename());
        fs::write(header_file, header)?;

        let c_file = bindings_path.join(config.c_filename());
        fs::write(c_file, c_file_content)?;

        if self.try_format_code {
            match Command::new("go").arg("fmt").arg(&go_file).output() {
                Ok(out) => {
                    if !out.status.success() {
                        let msg = match String::from_utf8(out.stderr) {
                            Ok(v) => v,
                            Err(e) => format!("{}", e).to_owned(),
                        };
                        println!(
                            "Warning: Unable to auto-format {} using go fmt: {}",
                            go_file.file_name().unwrap(),
                            msg
                        )
                    }
                }
                Err(e) => {
                    println!(
                        "Warning: Unable to auto-format {} using go fmt: {}",
                        go_file.file_name().unwrap(),
                        e
                    )
                }
            }
        }

        Ok(())
    }
}

fn full_bindings_path(config: &Config, ci: &ComponentInterface, out_dir: &Utf8Path) -> Utf8PathBuf {
    let package_path: Utf8PathBuf = config.package_name().split('.').collect();
    Utf8PathBuf::from(out_dir)
        .join(package_path)
        .join(ci.namespace())
}

pub fn main() -> anyhow::Result<()> {
    let Cli {
        out_dir,
        no_format,
        config,
        lib_file,
        library_mode,
        crate_name,
        source,
    } = Cli::parse();

    let binding_gen = BindingGeneratorGo {
        try_format_code: !no_format,
    };
    if library_mode {
        if lib_file.is_some() {
            panic!("--lib-file is not compatible with --library.")
        }
        let out_dir = out_dir.expect("--out-dir is required when using --library");
        let library_path = source;

        let cargo_metadata = MetadataCommand::new()
            .exec()
            .context("error running cargo metadata")?;
        let cdylib_name = uniffi_bindgen::library_mode::calc_cdylib_name(&library_path);
        let mut sources = find_sources(
            &cargo_metadata,
            &library_path,
            cdylib_name,
            config.as_deref(),
        )?;
        for i in 0..sources.len() {
            // Partition up the sources list because we're eventually going to call
            // `update_from_dependency_configs()` which requires an exclusive reference to one source and
            // shared references to all other sources.
            let (sources_before, rest) = sources.split_at_mut(i);
            let (source, sources_after) = rest.split_first_mut().unwrap();
            let other_sources = sources_before.iter().chain(sources_after.iter());
            // Calculate which configs come from dependent crates
            let dependencies = HashSet::<&str>::from_iter(
                source.package.dependencies.iter().map(|d| d.name.as_str()),
            );
            let config_map: HashMap<&str, &Config> = other_sources
                .filter_map(|s| {
                    dependencies
                        .contains(s.package.name.as_str())
                        .then_some((s.crate_name.as_str(), &s.config))
                })
                .collect();
            // We can finally call update_from_dependency_configs
            source.config.update_from_dependency_configs(config_map);
        }
        fs::create_dir_all(&out_dir)?;
        if let Some(crate_name) = &crate_name {
            let old_elements = sources.drain(..);
            let mut matches: Vec<_> = old_elements
                .filter(|s| &s.crate_name == crate_name)
                .collect();
            match matches.len() {
                0 => bail!("Crate {crate_name} not found in {library_path}"),
                1 => sources.push(matches.pop().unwrap()),
                n => bail!("{n} crates named {crate_name} found in {library_path}"),
            }
        }

        for source in sources.into_iter() {
            binding_gen.write_bindings(source.ci, source.config, &out_dir)?;
        }
    } else {
        uniffi_bindgen::generate_external_bindings(binding_gen, source, config, out_dir, lib_file)?;
    }

    Ok(())
}

// Copied from library_mode

fn find_sources(
    cargo_metadata: &cargo_metadata::Metadata,
    library_path: &Utf8Path,
    cdylib_name: Option<&str>,
    config_file_override: Option<&Utf8Path>,
) -> anyhow::Result<Vec<Source>> {
    uniffi_meta::group_metadata(macro_metadata::extract_from_library(library_path)?)?
        .into_iter()
        .map(|group| {
            let package = find_package_by_crate_name(cargo_metadata, &group.namespace.crate_name)?;
            let crate_root = package
                .manifest_path
                .parent()
                .context("manifest path has no parent")?;
            let crate_name = group.namespace.crate_name.clone();
            let mut ci = ComponentInterface::new(&crate_name);
            if let Some(metadata) = load_udl_metadata(&group, crate_root, &crate_name)? {
                ci.add_metadata(metadata)?;
            };
            ci.add_metadata(group)?;
            let mut config = Config::load_initial(crate_root, config_file_override)?;
            if let Some(cdylib_name) = cdylib_name {
                config.update_from_cdylib_name(cdylib_name);
            }
            config.update_from_ci(&ci);
            Ok(Source {
                config,
                crate_name,
                ci,
                package,
            })
        })
        .collect()
}

fn find_package_by_crate_name(
    metadata: &cargo_metadata::Metadata,
    crate_name: &str,
) -> anyhow::Result<Package> {
    let matching: Vec<&Package> = metadata
        .packages
        .iter()
        .filter(|p| {
            p.targets
                .iter()
                .any(|t| t.name.replace('-', "_") == crate_name)
        })
        .collect();
    match matching.len() {
        1 => Ok(matching[0].clone()),
        n => bail!("cargo metadata returned {n} packages for crate name {crate_name}"),
    }
}

fn load_udl_metadata(
    group: &uniffi_meta::MetadataGroup,
    crate_root: &Utf8Path,
    crate_name: &str,
) -> anyhow::Result<Option<uniffi_meta::MetadataGroup>> {
    let udl_items = group
        .items
        .iter()
        .filter_map(|i| match i {
            uniffi_meta::Metadata::UdlFile(meta) => Some(meta),
            _ => None,
        })
        .collect::<Vec<_>>();
    match udl_items.len() {
        // No UDL files, load directly from the group
        0 => Ok(None),
        // Found a UDL file, use it to load the CI, then add the MetadataGroup
        1 => {
            if udl_items[0].module_path != crate_name {
                bail!(
                    "UDL is for crate '{}' but this crate name is '{}'",
                    udl_items[0].module_path,
                    crate_name
                );
            }
            let ci_name = &udl_items[0].file_stub;
            let ci_path = crate_root.join("src").join(format!("{ci_name}.udl"));
            if ci_path.exists() {
                let udl = fs::read_to_string(ci_path)?;
                let udl_group = uniffi_udl::parse_udl(&udl, crate_name)?;
                Ok(Some(udl_group))
            } else {
                bail!("{ci_path} not found");
            }
        }
        n => bail!("{n} UDL files found for {crate_root}"),
    }
}

// A single source that we generate bindings for
#[derive(Debug)]
pub struct Source {
    pub package: Package,
    pub crate_name: String,
    pub ci: ComponentInterface,
    pub config: Config,
}
