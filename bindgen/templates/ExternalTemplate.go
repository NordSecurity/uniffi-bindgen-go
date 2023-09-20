{%- let mod_name = module_path|import_name %}
{%- let local_buffer_name = "RustBuffer{}"|format(name) %}

{{ self.add_local_import(mod_name) }}

type {{ local_buffer_name }} = {{ mod_name }}.RustBuffer
