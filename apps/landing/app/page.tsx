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
    <div className="relative overflow-hidden selection:bg-accent selection:text-background">
      {/* Ambient background blur */}
      <div className="pointer-events-none absolute left-[-20%] top-[-10%] h-125 w-125 rounded-full bg-accent/5 blur-[120px]" />
      <div className="pointer-events-none absolute right-[-20%] bottom-[-10%] h-125 w-125 rounded-full bg-primary/5 blur-[140px]" />

      <Navbar />

      <main className="mx-auto max-w-7xl px-6 sm:px-8">
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
