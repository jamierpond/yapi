import Navbar from "../components/Navbar";
import Playground from "../components/Playground";
import ShareButton from "../components/ShareButton";

export default function PlaygroundPage() {
  return (
    <div className="flex flex-col h-screen">
      <Navbar rightContent={<ShareButton />} />
      <Playground />
    </div>
  );
}
