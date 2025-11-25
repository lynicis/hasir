"use client";

import { Database, LayoutDashboard, LogIn, User, UserPlus } from "lucide-react";
import { motion, Variants } from "framer-motion";
import Link from "next/link";

import { useUserStore } from "@/stores/user-store-provider";
import { Button } from "@/components/ui/button";

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.15,
      delayChildren: 0.2,
    },
  },
};

const itemVariants: Variants = {
  hidden: { y: 30, opacity: 0 },
  visible: {
    y: 0,
    opacity: 1,
    transition: {
      duration: 0.6,
      ease: "easeOut",
    },
  },
};

export default function HomePage() {
  const { id, email } = useUserStore((state) => state);

  const isLoggedIn = Boolean(id && email);

  return (
    <div className="min-h-screen bg-background">
      <section className="relative flex min-h-[90vh] w-full items-center justify-center overflow-hidden px-6 py-20">
        <div className="absolute inset-0 bg-linear-to-br from-primary/5 via-background to-secondary/5" />
        <motion.div
          className="absolute top-20 right-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl"
          animate={{
            scale: [1, 1.2, 1],
            opacity: [0.3, 0.5, 0.3],
          }}
          transition={{
            duration: 8,
            repeat: Infinity,
            ease: "easeInOut",
          }}
        />
        <motion.div
          className="absolute bottom-20 left-20 w-96 h-96 bg-secondary/10 rounded-full blur-3xl"
          animate={{
            scale: [1, 1.3, 1],
            opacity: [0.2, 0.4, 0.2],
          }}
          transition={{
            duration: 10,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 1,
          }}
        />
        <motion.div
          className="relative z-10 flex max-w-5xl flex-col items-center text-center"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
        >
          <motion.div
            variants={itemVariants}
            className="mb-6 flex items-center justify-center"
          >
            <div className="rounded-2xl bg-primary/10 p-4 backdrop-blur-sm border border-primary/20">
              <Database className="w-12 h-12 text-primary" />
            </div>
          </motion.div>

          <motion.h1
            className="text-5xl font-bold tracking-tight sm:text-6xl md:text-7xl lg:text-8xl bg-linear-to-r from-foreground via-foreground to-foreground/60 bg-clip-text text-transparent"
            variants={itemVariants}
          >
            Hasir
          </motion.h1>
          <motion.p
            className="mt-6 max-w-2xl text-lg leading-8 text-muted-foreground md:text-xl"
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
                <Button asChild size="lg" className="min-w-[160px]">
                  <Link href="/dashboard" className="flex items-center gap-2">
                    <LayoutDashboard className="w-4 h-4" />
                    Dashboard
                  </Link>
                </Button>
                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="min-w-[160px]"
                >
                  <Link href="/profile" className="flex items-center gap-2">
                    <User className="w-4 h-4" />
                    Profile
                  </Link>
                </Button>
              </>
            ) : (
              <>
                <Button asChild size="lg" className="min-w-[160px]">
                  <Link href="/login" className="flex items-center gap-2">
                    <LogIn className="w-4 h-4" />
                    Login
                  </Link>
                </Button>
                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="min-w-[160px]"
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
              className="mt-6 text-sm text-muted-foreground"
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
