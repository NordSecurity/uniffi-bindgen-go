{%- let namespace = ci.namespace_for_type(type_).expect("external type should have namespace") %}
{%- let ns = namespace|import_name %}
{{ self.add_local_import(ns) }}
