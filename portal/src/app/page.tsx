// Portal root route ("/"). Renders an Overview placeholder until the
// dashboard is implemented.
import { PlaceholderPage } from "./components/PlaceholderPage";

export default function HomePage() {
  return (
    <PlaceholderPage
      title="Overview"
      description="A high-level summary of your portal activity will live here."
    />
  );
}
