import Link from "next/link";
import { FC } from "react";

const Header: FC = () => {
  return (
    <header className="bg-green-950 shadow-md">
      <nav className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex">
            <Link href="/" className="flex-shrink-0 flex items-center">
              <span className="text-2xl font-bold text-white">
                The BrYan Alchemind
              </span>
            </Link>
            <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
              <Link
                href="/about"
                className="border-b-2 border-transparent text-shadow-white hover:border-white hover:text-emerald-500 inline-flex items-center px-1 pt-1 text-base font-semibold"
              >
                About
              </Link>
              <Link
                href="/game"
                className="border-b-2 border-transparent text-shadow-white hover:border-white hover:text-emerald-500 inline-flex items-center px-1 pt-1 text-base font-semibold"
              >
                Game
              </Link>
            </div>
          </div>
        </div>
      </nav>
    </header>
  );
};

export default Header;
