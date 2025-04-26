<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Carbon;

/**
 * App\Models\Job
 * @property string $id
 * @property string $title
 * @property string $description
 * @property string $source
 * @property string $location
 * @property string $url
 * @property bool $remote
 * @property Carbon $posted_at
 * @property Carbon $created_at
 * @property Carbon $updated_at
 * @property Carbon $deleted_at
 */
class Job extends Model
{
    use HasFactory;

    protected $guarded = [];
}
