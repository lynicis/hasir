"use client";

import { LayoutDashboard, LogIn, User, UserPlus } from "lucide-react";
import { motion, type Variants } from "framer-motion";
import Image from "next/image";
import Link from "next/link";

import { useSession } from "@/lib/session-provider";
import { Button } from "@/components/ui/button";

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.1,
    },
  },
};

const itemVariants: Variants = {
  hidden: { y: 20, opacity: 0 },
  visible: {
    y: 0,
    opacity: 1,
    transition: {
      type: "spring",
      stiffness: 100,
      damping: 15,
    },
  },
};

export default function HomePageContent() {
  const { session } = useSession();
  const isLoggedIn = Boolean(session?.user?.id);
  const email = session?.user?.email;

  return (
    <div className="min-h-screen bg-background relative overflow-hidden noise-bg">
      {/* Dynamic ambient spotlight backdrops */}
      <div className="absolute top-0 left-1/4 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-10 right-1/4 w-[600px] h-[600px] bg-secondary/5 rounded-full blur-[140px] pointer-events-none" />
      
      <section className="relative flex min-h-[90vh] w-full items-center justify-center px-6 py-24">
        <motion.div
          className="relative z-10 flex max-w-4xl flex-col items-center text-center"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
        >
          <motion.div
            variants={itemVariants}
            className="mb-8"
          >
            <div className="rounded-xl bg-card border p-4 relative group hover:border-primary/45 transition-colors duration-500">
              <div className="absolute inset-0 bg-primary/5 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              <Image
                src="/logo.svg"
                alt="Hasir Logo"
                width={40}
                height={40}
                className="relative z-10 hidden dark:block"
                priority
              />
              <Image
                src="/logo-light.svg"
                alt="Hasir Logo"
                width={40}
                height={40}
                className="relative z-10 block dark:hidden"
                priority
              />
            </div>
          </motion.div>

          <motion.h1
            className="text-6xl font-extrabold tracking-tighter sm:text-7xl md:text-8xl leading-none text-foreground"
            variants={itemVariants}
          >
            Hasir
          </motion.h1>
          <motion.p
            className="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg text-pretty"
            variants={itemVariants}
          >
            Your modern Proto Schema Registry dashboard. Manage, version, and
            collaborate on your protocol buffer schemas with ease.
          </motion.p>
          <motion.div
            className="mt-10 flex flex-col sm:flex-row items-center gap-4"
            variants={itemVariants}
          >
            {isLoggedIn ? (
              <>
                <Button 
                  asChild 
                  size="lg" 
                  className="min-w-[160px] active:scale-[0.97] active:translate-y-[1px] transition-transform duration-100"
                >
                  <Link href="/dashboard" className="flex items-center gap-2">
                    <LayoutDashboard className="w-4 h-4" />
                    Dashboard
                  </Link>
                </Button>
                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="min-w-[160px] active:scale-[0.97] active:translate-y-[1px] transition-transform duration-100"
                >
                  <Link href="/profile" className="flex items-center gap-2">
                    <User className="w-4 h-4" />
                    Profile
                  </Link>
                </Button>
              </>
            ) : (
              <>
                <Button 
                  asChild 
                  size="lg" 
                  className="min-w-[160px] active:scale-[0.97] active:translate-y-[1px] transition-transform duration-100"
                >
                  <Link href="/login" className="flex items-center gap-2">
                    <LogIn className="w-4 h-4" />
                    Login
                  </Link>
                </Button>
                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="min-w-[160px] active:scale-[0.97] active:translate-y-[1px] transition-transform duration-100"
                >
                  <Link href="/register" className="flex items-center gap-2">
                    <UserPlus className="w-4 h-4" />
                    Register
                  </Link>
                </Button>
              </>
            )}
          </motion.div>
          {isLoggedIn && (
            <motion.p
              className="mt-6 text-xs tracking-wide text-muted-foreground font-mono"
              variants={itemVariants}
            >
              Welcome back,{" "}
              <span className="font-semibold text-foreground">{email}</span>!
            </motion.p>
          )}
        </motion.div>
      </section>
    </div>
  );
}
