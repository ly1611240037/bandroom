export default function Field({ label, name, type = "text", value, onChange, required, autoComplete, ...props }) {
  return <label className="field"><span>{label}</span><input name={name} type={type} value={value} onChange={(event) => onChange((old) => ({ ...old, [event.target.name]: event.target.value }))} required={required} autoComplete={autoComplete || (type === "password" ? "new-password" : name === "phone" ? "tel" : name)} {...props} /></label>;
}
