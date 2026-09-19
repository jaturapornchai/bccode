import { NextResponse } from "next/server";
import { askKmsChatbot, type ChatMessage } from "@/lib/kms-rag";

export async function POST(req: Request) {
  try {
    const body = await req.json();
    const messages = body.messages as ChatMessage[];

    if (!Array.isArray(messages) || messages.length === 0) {
      return NextResponse.json(
        { error: "Invalid request payload: messages array required" },
        { status: 400 }
      );
    }

    const { answer, sources } = await askKmsChatbot(messages);

    return NextResponse.json({
      success: true,
      answer,
      sources,
    });
  } catch (error: any) {
    console.error("KMS Chatbot API error:", error);
    return NextResponse.json(
      {
        success: false,
        error: error?.message || "Internal server error occurred while contacting AI service",
      },
      { status: 500 }
    );
  }
}
