"use client";

import { LanguagesSection } from "@/components/language";
import { WorkflowSection } from "@/components/workflow";
import { FeaturesSection } from "@/components/feature";
import { BottomCTA } from "@/components/bottom-cta";
import { HeroSection } from "@/components/hero";
import { Footer } from "@/components/footer";
import { Navbar } from "@/components/navbar";

export default function LandingPage() {
  return (
    <div>
      <a
        className="skip focus-visible:outline-2 focus-visible:outline-focus active:translate-y-px"
        href="#main"
      >
        Skip to content
      </a>

      <Navbar />

      <main id="main">
        <HeroSection />
        <FeaturesSection />
        <WorkflowSection />
        <LanguagesSection />
        <BottomCTA />
      </main>

      <Footer />
    </div>
  );
}
