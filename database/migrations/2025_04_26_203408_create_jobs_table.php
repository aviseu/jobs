<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('jobs', function (Blueprint $table) {
            $table->string('id')->primary();
            $table->string('channel_id');
            $table->string('title');
            $table->string('description');
            $table->string('source');
            $table->string('location');
            $table->string('url');
            $table->boolean('remote');
            $table->timestampTz('posted_at');
            $table->timestamps();
            $table->softDeletes();

            $table->index('id', 'idx_jobs_id');
            $table->index('posted_at', 'idx_jobs_posted_at');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('jobs');
    }
};
